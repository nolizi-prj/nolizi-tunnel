package mux

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/pumasi-ai/pumasi-tunnel/core"
)

// pair returns two sessions joined by an in-memory connection: the agent side
// (client, odd ids) and the relay side (server, even ids).
func pair(t *testing.T) (client, server *Session) {
	t.Helper()
	c, s := net.Pipe()
	client = Client(c)
	server = Server(s)
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})
	return client, server
}

func TestStreamRoundTrip(t *testing.T) {
	client, server := pair(t)

	done := make(chan error, 1)
	go func() {
		st, err := server.Accept()
		if err != nil {
			done <- err
			return
		}
		defer st.Close()
		got, err := io.ReadAll(st)
		if err != nil {
			done <- err
			return
		}
		if string(got) != "hello relay" {
			done <- fmt.Errorf("server read %q", got)
			return
		}
		_, err = st.Write([]byte("hello agent"))
		done <- err
	}()

	st, err := client.Open(0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := st.Write([]byte("hello relay")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	// Half-close: the server's ReadAll must finish while this side still reads.
	if err := st.CloseWrite(); err != nil {
		t.Fatalf("CloseWrite: %v", err)
	}

	reply := make([]byte, 64)
	n, err := st.Read(reply)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(reply[:n]) != "hello agent" {
		t.Errorf("client read %q, want %q", reply[:n], "hello agent")
	}
	if err := <-done; err != nil {
		t.Fatalf("server side: %v", err)
	}
}

// Stream ids are partitioned so both ends can open at once without collision.
func TestStreamIDsDoNotCollide(t *testing.T) {
	client, server := pair(t)

	a, err := client.Open(0)
	if err != nil {
		t.Fatalf("client Open: %v", err)
	}
	b, err := server.Open(0)
	if err != nil {
		t.Fatalf("server Open: %v", err)
	}
	if a.ID()%2 != 1 {
		t.Errorf("client stream id %d is not odd", a.ID())
	}
	if b.ID()%2 != 0 {
		t.Errorf("server stream id %d is not even", b.ID())
	}
}

func TestConcurrentStreamsAreIndependent(t *testing.T) {
	client, server := pair(t)

	// Echo every stream the client opens.
	go func() {
		for {
			st, err := server.Accept()
			if err != nil {
				return
			}
			go func(st *Stream) {
				defer st.Close()
				io.Copy(st, st)
				st.CloseWrite()
			}(st)
		}
	}()

	const streams = 12
	var wg sync.WaitGroup
	errs := make(chan error, streams)
	for i := 0; i < streams; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st, err := client.Open(0)
			if err != nil {
				errs <- err
				return
			}
			defer st.Close()

			want := []byte(fmt.Sprintf("payload for stream %d", i))
			if _, err := st.Write(want); err != nil {
				errs <- err
				return
			}
			st.CloseWrite()

			got, err := io.ReadAll(st)
			if err != nil {
				errs <- err
				return
			}
			if !bytes.Equal(got, want) {
				errs <- fmt.Errorf("stream %d: got %q, want %q", i, got, want)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// Payloads larger than one frame must be split on write and rejoined on read
// without corruption — this is what a file download through a tunnel does.
func TestLargePayloadIsChunkedAndReassembled(t *testing.T) {
	client, server := pair(t)

	payload := make([]byte, core.MaxPayloadSize*2+1234)
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand: %v", err)
	}

	got := make(chan []byte, 1)
	errCh := make(chan error, 1)
	go func() {
		st, err := server.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer st.Close()
		data, err := io.ReadAll(st)
		if err != nil {
			errCh <- err
			return
		}
		got <- data
	}()

	st, err := client.Open(0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	go func() {
		if _, err := st.Write(payload); err != nil {
			errCh <- err
			return
		}
		st.CloseWrite()
	}()

	select {
	case data := <-got:
		if !bytes.Equal(data, payload) {
			t.Errorf("reassembled %d bytes, want %d (equal: %v)", len(data), len(payload), bytes.Equal(data, payload))
		}
	case err := <-errCh:
		t.Fatalf("transfer: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatal("timed out reassembling a large payload")
	}
}

// A reader must see every buffered byte before EOF, even when the writer
// closes immediately after writing.
func TestBufferedDataDrainedBeforeEOF(t *testing.T) {
	client, server := pair(t)

	go func() {
		st, err := client.Open(0)
		if err != nil {
			return
		}
		st.Write([]byte("first"))
		st.Write([]byte("second"))
		st.CloseWrite()
	}()

	st, err := server.Accept()
	if err != nil {
		t.Fatalf("Accept: %v", err)
	}
	defer st.Close()

	got, err := io.ReadAll(st)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "firstsecond" {
		t.Errorf("got %q, want %q", got, "firstsecond")
	}
}

// The distinction that keeps a truncated body from looking complete: a peer
// that finished writing yields io.EOF, but a session that dies mid-stream
// must yield an error instead.
func TestSessionDeathIsNotEOF(t *testing.T) {
	client, server := pair(t)

	go func() {
		st, err := client.Open(0)
		if err != nil {
			return
		}
		st.Write([]byte("partial body"))
		// No CloseWrite: the peer never says it is finished.
		time.Sleep(50 * time.Millisecond)
		client.Close() // connection dies mid-stream
	}()

	st, err := server.Accept()
	if err != nil {
		t.Fatalf("Accept: %v", err)
	}
	defer st.Close()

	_, err = io.ReadAll(st)
	if err == nil {
		t.Fatal("ReadAll returned nil error after the session died mid-stream")
	}
	if !errors.Is(err, ErrSessionClosed) {
		t.Errorf("got %v, want ErrSessionClosed", err)
	}
}

// By contrast, a clean end-of-stream followed by teardown is a real EOF.
func TestCleanEndOfStreamIsEOF(t *testing.T) {
	client, server := pair(t)

	go func() {
		st, err := client.Open(0)
		if err != nil {
			return
		}
		st.Write([]byte("complete body"))
		st.CloseWrite()
	}()

	st, err := server.Accept()
	if err != nil {
		t.Fatalf("Accept: %v", err)
	}
	defer st.Close()

	got, err := io.ReadAll(st)
	if err != nil {
		t.Fatalf("ReadAll after a clean close: %v", err)
	}
	if string(got) != "complete body" {
		t.Errorf("got %q", got)
	}
}

func TestOperationsAfterCloseFail(t *testing.T) {
	client, server := pair(t)

	st, err := client.Open(0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// A second Close is a no-op so deferred closes are safe.
	if err := st.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
	if _, err := st.Write([]byte("x")); !errors.Is(err, ErrStreamClosed) {
		t.Errorf("Write after Close = %v, want ErrStreamClosed", err)
	}
	if _, err := st.Read(make([]byte, 4)); !errors.Is(err, ErrStreamClosed) {
		t.Errorf("Read after Close = %v, want ErrStreamClosed", err)
	}

	client.Close()
	if _, err := client.Open(0); !errors.Is(err, ErrSessionClosed) {
		t.Errorf("Open after session Close = %v, want ErrSessionClosed", err)
	}

	// The stream opened above reached the server before the close, and must
	// still be delivered: a request already in flight is not dropped just
	// because the connection ended afterwards.
	if _, err := server.Accept(); err != nil {
		t.Errorf("Accept should first drain the stream that arrived before the close: %v", err)
	}
	// Only once the backlog is empty does Accept report the dead session,
	// rather than blocking a serving loop forever.
	if _, err := server.Accept(); err == nil {
		t.Error("Accept on a drained dead session should fail, not block")
	}
}

// A serving loop must terminate when the peer disappears rather than hang.
func TestAcceptUnblocksOnPeerDeath(t *testing.T) {
	client, server := pair(t)

	done := make(chan error, 1)
	go func() {
		_, err := server.Accept()
		done <- err
	}()

	time.Sleep(20 * time.Millisecond)
	client.Close()

	select {
	case err := <-done:
		if err == nil {
			t.Error("Accept returned nil error after the peer died")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Accept did not unblock when the peer died")
	}
}

func TestNumStreamsTracksLifecycle(t *testing.T) {
	client, _ := pair(t)

	if got := client.NumStreams(); got != 0 {
		t.Fatalf("NumStreams = %d on a fresh session", got)
	}
	a, err := client.Open(0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got := client.NumStreams(); got != 1 {
		t.Errorf("NumStreams = %d after one Open, want 1", got)
	}
	a.Close()
	if got := client.NumStreams(); got != 0 {
		t.Errorf("NumStreams = %d after Close, want 0", got)
	}
}

// The TCP flag must survive to the accepting side, which is how the relay
// knows to forward bytes without parsing HTTP.
func TestTCPFlagCrossesTheWire(t *testing.T) {
	client, server := pair(t)

	go func() {
		st, err := client.Open(core.FlagTCP)
		if err == nil {
			st.Write([]byte("raw"))
		}
	}()

	st, err := server.Accept()
	if err != nil {
		t.Fatalf("Accept: %v", err)
	}
	defer st.Close()
	if !st.IsTCP() {
		t.Error("accepted stream did not carry the TCP flag")
	}
}

// A relay under load opens and closes streams constantly; leaking one per
// request would exhaust the map over a long-lived session.
func TestNoStreamLeakUnderChurn(t *testing.T) {
	client, server := pair(t)

	go func() {
		for {
			st, err := server.Accept()
			if err != nil {
				return
			}
			st.Close()
		}
	}()

	for i := 0; i < 200; i++ {
		st, err := client.Open(0)
		if err != nil {
			t.Fatalf("Open %d: %v", i, err)
		}
		if err := st.Close(); err != nil {
			t.Fatalf("Close %d: %v", i, err)
		}
	}
	if got := client.NumStreams(); got != 0 {
		t.Errorf("client leaked %d streams", got)
	}
}

func BenchmarkStreamThroughput(b *testing.B) {
	c, s := net.Pipe()
	client, server := Client(c), Server(s)
	defer client.Close()
	defer server.Close()

	go func() {
		for {
			st, err := server.Accept()
			if err != nil {
				return
			}
			go func(st *Stream) {
				io.Copy(io.Discard, st)
				st.Close()
			}(st)
		}
	}()

	payload := make([]byte, 32*1024)
	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()
	b.ResetTimer()

	st, err := client.Open(0)
	if err != nil {
		b.Fatal(err)
	}
	defer st.Close()
	for i := 0; i < b.N; i++ {
		if _, err := st.Write(payload); err != nil {
			b.Fatal(err)
		}
	}
}
