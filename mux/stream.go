package mux

import (
	"io"
	"sync"

	"github.com/pumasi-ai/pumasi-tunnel/core"
)

// Stream is one logical connection inside a session. It satisfies
// io.ReadWriteCloser, so the relay and the agent can hand it to io.Copy and
// to net/http without either knowing a multiplexer is underneath.
//
// Half-close is real: Read returns io.EOF once the peer finishes writing,
// while Write keeps working until this side closes. That is what lets an
// HTTP response body flow back after the request body has ended.
type Stream struct {
	id      uint32
	flags   uint8
	session *Session

	recv     chan []byte
	leftover []byte

	// localClosed is closed by Close; remoteDone by an end-of-stream or close
	// frame from the peer. Both are one-shot, guarded by their sync.Once.
	localClosed chan struct{}
	remoteDone  chan struct{}
	closeOnce   sync.Once
	remoteOnce  sync.Once
}

func newStream(id uint32, s *Session) *Stream {
	return &Stream{
		id:          id,
		session:     s,
		recv:        make(chan []byte, streamBufferFrames),
		localClosed: make(chan struct{}),
		remoteDone:  make(chan struct{}),
	}
}

// ID reports the stream's protocol id, which appears in relay logs.
func (st *Stream) ID() uint32 { return st.id }

// IsTCP reports whether the peer opened this stream for raw TCP rather than
// HTTP, so the relay knows to skip host parsing.
func (st *Stream) IsTCP() bool { return st.flags&core.FlagTCP != 0 }

// Read returns bytes the peer wrote, blocking until some arrive. Buffered
// data is always drained before any end condition is reported, so a fast
// writer followed by an immediate close never loses its payload.
//
// The end condition matters: io.EOF is returned only when the peer actually
// finished writing. A session that dies mid-stream reports the session's
// error instead, because reporting EOF there would present a truncated
// response body to a visitor as a complete one.
func (st *Stream) Read(p []byte) (int, error) {
	if len(st.leftover) > 0 {
		return st.consume(p, nil), nil
	}
	// Buffered data outranks every close signal; without this first
	// non-blocking check, Go's random select choice could report the end of
	// the stream while bytes were still queued.
	select {
	case chunk := <-st.recv:
		return st.consume(p, chunk), nil
	default:
	}

	select {
	case chunk := <-st.recv:
		return st.consume(p, chunk), nil
	case <-st.remoteDone:
		return st.drainThen(p, io.EOF)
	case <-st.localClosed:
		return 0, ErrStreamClosed
	case <-st.session.closed:
		return st.drainThen(p, st.sessionEndErr())
	}
}

// consume copies from leftover, or from chunk when one was just received,
// stashing whatever does not fit.
func (st *Stream) consume(p []byte, chunk []byte) int {
	if chunk == nil {
		n := copy(p, st.leftover)
		st.leftover = st.leftover[n:]
		return n
	}
	n := copy(p, chunk)
	if n < len(chunk) {
		st.leftover = chunk[n:]
	}
	return n
}

// drainThen yields any frame still buffered before reporting end.
func (st *Stream) drainThen(p []byte, end error) (int, error) {
	select {
	case chunk := <-st.recv:
		return st.consume(p, chunk), nil
	default:
		return 0, end
	}
}

// sessionEndErr distinguishes "the peer finished, then the session ended"
// from "the session died with the stream still open".
func (st *Stream) sessionEndErr() error {
	select {
	case <-st.remoteDone:
		return io.EOF
	default:
		return st.session.closedError()
	}
}

// Write sends bytes to the peer, splitting anything larger than one frame.
func (st *Stream) Write(p []byte) (int, error) {
	select {
	case <-st.localClosed:
		return 0, ErrStreamClosed
	case <-st.session.closed:
		return 0, st.session.closedError()
	default:
	}

	written := 0
	for len(p) > 0 {
		chunk := p
		if len(chunk) > core.MaxPayloadSize {
			chunk = chunk[:core.MaxPayloadSize]
		}
		err := st.session.writeFrame(core.Frame{
			Type:     core.FrameData,
			StreamID: st.id,
			Payload:  chunk,
		})
		if err != nil {
			return written, err
		}
		written += len(chunk)
		p = p[len(chunk):]
	}
	return written, nil
}

// CloseWrite ends this side's half of the stream, letting the peer see EOF
// while this side keeps reading. An HTTP client uses it to signal the end of
// a request body.
func (st *Stream) CloseWrite() error {
	return st.session.writeFrame(core.Frame{
		Type:     core.FrameData,
		Flags:    core.FlagEndStream,
		StreamID: st.id,
	})
}

// Close ends the stream in both directions and releases it from the session.
// Closing twice is not an error, so deferred closes are safe.
func (st *Stream) Close() error {
	var err error
	st.closeOnce.Do(func() {
		close(st.localClosed)
		st.session.removeStream(st.id)
		// Best-effort: if the session is already gone the peer needs no
		// notice, and reporting that as this stream's failure would be noise.
		if e := st.session.writeFrame(core.Frame{Type: core.FrameClose, StreamID: st.id}); e != nil {
			select {
			case <-st.session.closed:
			default:
				err = e
			}
		}
	})
	return err
}

// deliver hands a payload to the stream's reader. It reports false if the
// stream or session ended first, so the read loop can drop the frame instead
// of blocking forever on a dead stream.
func (st *Stream) deliver(payload []byte) bool {
	select {
	case st.recv <- payload:
		return true
	case <-st.localClosed:
		return false
	case <-st.session.closed:
		return false
	}
}

// remoteClosed marks the peer as finished writing. Only a frame from the peer
// may call this: a dead session is not a clean end of stream, and conflating
// the two is what would truncate a body silently.
func (st *Stream) remoteClosed() {
	st.remoteOnce.Do(func() { close(st.remoteDone) })
}
