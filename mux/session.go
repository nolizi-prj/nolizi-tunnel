// Package mux carries many logical streams over one connection, using the
// frame protocol in core. It is the transport the whole product rests on: a
// client holds a single outbound connection to the relay, and every inbound
// visitor request becomes a stream travelling back down it.
//
// The split from core is deliberate — core decides what bytes mean and is
// pure; this package is the I/O shell that moves them.
//
// Flow control is per-stream buffering, not a credit window: a reader that
// stops reading eventually stalls its own stream and, once the read loop
// blocks handing off a frame, the connection behind it. That is honest
// backpressure rather than unbounded memory, but it does allow one stalled
// stream to hold up its siblings. A credit window is the fix and is deferred;
// see docs/ux/incumbent-ux-spec.md §9 item 1.
package mux

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/pumasi-ai/pumasi-tunnel/core"
)

var (
	ErrSessionClosed = errors.New("mux: session is closed")
	ErrStreamClosed  = errors.New("mux: stream is closed")
	ErrStreamExists  = errors.New("mux: peer reopened a live stream id")
)

// streamBufferFrames bounds how many payloads a stream may hold before the
// read loop stalls on it. Small on purpose: the buffer is a jitter absorber,
// not a queue.
const streamBufferFrames = 8

// Session multiplexes streams over one connection. It is safe for concurrent
// use by many goroutines.
type Session struct {
	conn io.ReadWriteCloser

	// Stream ids are partitioned by role so the two ends can open streams
	// simultaneously without colliding: the dialling side uses odd ids, the
	// accepting side even.
	odd bool

	writeMu sync.Mutex

	mu      sync.Mutex
	streams map[uint32]*Stream
	nextID  uint32

	accept    chan *Stream
	closed    chan struct{}
	closeOnce sync.Once

	errMu    sync.Mutex
	closeErr error
}

// Client returns a session that opens odd-numbered streams. The tunnel agent
// dials, so it is the client.
func Client(conn io.ReadWriteCloser) *Session { return newSession(conn, true) }

// Server returns a session that opens even-numbered streams. The relay
// accepts, so it is the server.
func Server(conn io.ReadWriteCloser) *Session { return newSession(conn, false) }

func newSession(conn io.ReadWriteCloser, odd bool) *Session {
	s := &Session{
		conn:    conn,
		odd:     odd,
		streams: make(map[uint32]*Stream),
		accept:  make(chan *Stream, 16),
		closed:  make(chan struct{}),
	}
	if odd {
		s.nextID = 1
	} else {
		s.nextID = 2
	}
	go s.readLoop()
	return s
}

// Open starts a new stream. flags may carry core.FlagTCP to tell the peer
// this stream is raw bytes rather than HTTP.
func (s *Session) Open(flags uint8) (*Stream, error) {
	s.mu.Lock()
	select {
	case <-s.closed:
		s.mu.Unlock()
		return nil, s.closedError()
	default:
	}
	id := s.nextID
	s.nextID += 2
	st := newStream(id, s)
	s.streams[id] = st
	s.mu.Unlock()

	if err := s.writeFrame(core.Frame{Type: core.FrameOpen, Flags: flags, StreamID: id}); err != nil {
		s.removeStream(id)
		return nil, err
	}
	return st, nil
}

// Accept returns the next stream opened by the peer. It returns the session's
// close error once the connection is gone, so a serving loop terminates.
func (s *Session) Accept() (*Stream, error) {
	select {
	case st := <-s.accept:
		return st, nil
	case <-s.closed:
		// Drain anything that arrived before the close was observed, so a
		// stream is never silently dropped in a race with teardown.
		select {
		case st := <-s.accept:
			return st, nil
		default:
		}
		return nil, s.closedError()
	}
}

// Close tears down the session and every stream on it.
func (s *Session) Close() error {
	return s.closeWith(nil)
}

// CloseChan is closed when the session ends; callers select on it to notice
// a dead peer without polling.
func (s *Session) CloseChan() <-chan struct{} { return s.closed }

// NumStreams reports how many streams are live, for the status surface.
func (s *Session) NumStreams() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.streams)
}

func (s *Session) closeWith(cause error) error {
	var err error
	s.closeOnce.Do(func() {
		s.errMu.Lock()
		s.closeErr = cause
		s.errMu.Unlock()

		// Closing s.closed is what wakes every blocked reader and writer:
		// each of them selects on it. Streams are deliberately not marked
		// remote-closed here — a dead session is not a clean end of stream,
		// and Stream.Read reports the difference.
		close(s.closed)
		err = s.conn.Close()

		s.mu.Lock()
		s.streams = map[uint32]*Stream{}
		s.mu.Unlock()
	})
	return err
}

func (s *Session) closedError() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	if s.closeErr != nil {
		return fmt.Errorf("%w: %v", ErrSessionClosed, s.closeErr)
	}
	return ErrSessionClosed
}

// writeFrame serialises one frame onto the connection. Writes from many
// streams interleave, so the mutex is what keeps frames from being spliced
// into each other.
func (s *Session) writeFrame(f core.Frame) error {
	wire, err := f.Encode()
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	select {
	case <-s.closed:
		return s.closedError()
	default:
	}
	if _, err := s.conn.Write(wire); err != nil {
		s.closeWith(err)
		return err
	}
	return nil
}

func (s *Session) removeStream(id uint32) {
	s.mu.Lock()
	delete(s.streams, id)
	s.mu.Unlock()
}

func (s *Session) readLoop() {
	for {
		f, err := core.DecodeFrame(s.conn)
		if err != nil {
			if err == io.EOF {
				err = nil // the peer hung up cleanly
			}
			s.closeWith(err)
			return
		}

		switch f.Type {
		case core.FrameOpen:
			s.handleOpen(f)

		case core.FrameData:
			s.mu.Lock()
			st := s.streams[f.StreamID]
			s.mu.Unlock()
			if st == nil {
				// Data for a stream we already closed: the peer had frames in
				// flight. Dropping is correct, not an error.
				continue
			}
			if len(f.Payload) > 0 {
				if !st.deliver(f.Payload) {
					continue
				}
			}
			if f.Flags&core.FlagEndStream != 0 {
				st.remoteClosed()
			}

		case core.FrameClose:
			s.mu.Lock()
			st := s.streams[f.StreamID]
			s.mu.Unlock()
			if st != nil {
				st.remoteClosed()
			}

		case core.FramePing:
			// Reply on the same payload so the peer can match probes.
			_ = s.writeFrame(core.Frame{Type: core.FramePong, StreamID: f.StreamID, Payload: f.Payload})

		case core.FramePong, core.FrameAuth, core.FrameAuthOK, core.FrameError:
			// Control frames are the relay's and agent's business; the
			// multiplexer carries them and does not interpret them. A future
			// control channel hook belongs here.
		}
	}
}

func (s *Session) handleOpen(f core.Frame) {
	s.mu.Lock()
	if _, exists := s.streams[f.StreamID]; exists {
		s.mu.Unlock()
		// A peer reusing a live id is a protocol violation; tearing the
		// session down is safer than serving two streams on one id.
		s.closeWith(fmt.Errorf("%w: %d", ErrStreamExists, f.StreamID))
		return
	}
	st := newStream(f.StreamID, s)
	st.flags = f.Flags
	s.streams[f.StreamID] = st
	s.mu.Unlock()

	select {
	case s.accept <- st:
	case <-s.closed:
	}
}
