package core

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/quick"
)

func TestFrameRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		frame Frame
	}{
		{"open http", Frame{Type: FrameOpen, StreamID: 1}},
		{"open tcp", Frame{Type: FrameOpen, Flags: FlagTCP, StreamID: 7}},
		{"data", Frame{Type: FrameData, StreamID: 42, Payload: []byte("GET / HTTP/1.1\r\n\r\n")}},
		{"data end of stream", Frame{Type: FrameData, Flags: FlagEndStream, StreamID: 42, Payload: []byte("tail")}},
		{"close", Frame{Type: FrameClose, StreamID: 42}},
		{"ping with payload", Frame{Type: FramePing, Payload: []byte{0xDE, 0xAD}}},
		{"auth", Frame{Type: FrameAuth, Payload: []byte("token=abc;name=myapi")}},
		{"error", Frame{Type: FrameError, Payload: []byte("subdomain taken")}},
		{"max stream id", Frame{Type: FrameData, StreamID: ^uint32(0), Payload: []byte("x")}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := tc.frame.Encode()
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if len(wire) != HeaderSize+len(tc.frame.Payload) {
				t.Fatalf("encoded length = %d, want %d", len(wire), HeaderSize+len(tc.frame.Payload))
			}

			got, err := DecodeFrame(bytes.NewReader(wire))
			if err != nil {
				t.Fatalf("DecodeFrame: %v", err)
			}
			if got.Type != tc.frame.Type || got.Flags != tc.frame.Flags || got.StreamID != tc.frame.StreamID {
				t.Errorf("header round-trip: got %+v, want %+v", got, tc.frame)
			}
			if !bytes.Equal(got.Payload, tc.frame.Payload) {
				t.Errorf("payload round-trip: got %q, want %q", got.Payload, tc.frame.Payload)
			}
		})
	}
}

// The codec must survive arbitrary payloads, not just the ones we thought of.
func TestFrameRoundTripQuick(t *testing.T) {
	f := func(flags uint8, streamID uint32, payload []byte) bool {
		if len(payload) > MaxPayloadSize {
			return true
		}
		original := Frame{Type: FrameData, Flags: flags, StreamID: streamID, Payload: payload}
		wire, err := original.Encode()
		if err != nil {
			return false
		}
		got, err := DecodeFrame(bytes.NewReader(wire))
		if err != nil {
			return false
		}
		return got.Type == FrameData && got.Flags == flags && got.StreamID == streamID &&
			bytes.Equal(got.Payload, payload)
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}

// Frames must decode back-to-back from one stream: this is how the relay
// actually reads a connection.
func TestDecodeFrameSequence(t *testing.T) {
	want := []Frame{
		{Type: FrameAuth, Payload: []byte("token")},
		{Type: FrameOpen, StreamID: 1},
		{Type: FrameData, StreamID: 1, Payload: []byte("hello")},
		{Type: FrameData, StreamID: 1, Flags: FlagEndStream, Payload: []byte("world")},
		{Type: FrameClose, StreamID: 1},
	}

	var buf bytes.Buffer
	for _, f := range want {
		wire, err := f.Encode()
		if err != nil {
			t.Fatalf("Encode %v: %v", f.Type, err)
		}
		buf.Write(wire)
	}

	for i, expected := range want {
		got, err := DecodeFrame(&buf)
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if got.Type != expected.Type || got.StreamID != expected.StreamID {
			t.Errorf("frame %d: got %v/%d, want %v/%d", i, got.Type, got.StreamID, expected.Type, expected.StreamID)
		}
	}
	if _, err := DecodeFrame(&buf); err != io.EOF {
		t.Errorf("after the last frame: got %v, want io.EOF", err)
	}
}

// A clean close between frames is EOF; a severed connection mid-frame is not.
// The relay distinguishes these to decide whether to log an error.
func TestDecodeFrameTruncation(t *testing.T) {
	full, err := Frame{Type: FrameData, StreamID: 3, Payload: []byte("0123456789")}.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	t.Run("empty reader is clean EOF", func(t *testing.T) {
		if _, err := DecodeFrame(bytes.NewReader(nil)); err != io.EOF {
			t.Errorf("got %v, want io.EOF", err)
		}
	})

	t.Run("partial header", func(t *testing.T) {
		_, err := DecodeFrame(bytes.NewReader(full[:HeaderSize-1]))
		if !errors.Is(err, ErrShortHeader) {
			t.Errorf("got %v, want ErrShortHeader", err)
		}
	})

	t.Run("header without payload", func(t *testing.T) {
		_, err := DecodeFrame(bytes.NewReader(full[:HeaderSize]))
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("got %v, want io.ErrUnexpectedEOF", err)
		}
	})

	t.Run("partial payload", func(t *testing.T) {
		_, err := DecodeFrame(bytes.NewReader(full[:len(full)-3]))
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("got %v, want io.ErrUnexpectedEOF", err)
		}
	})
}

// A forged length prefix must be refused from the header alone — before any
// allocation — or a single hostile peer can exhaust the relay's memory.
func TestDecodeFrameRejectsOversizedLengthWithoutAllocating(t *testing.T) {
	header := make([]byte, HeaderSize)
	header[0] = uint8(FrameData)
	// Declare 4 GiB while supplying no payload at all.
	header[6], header[7], header[8], header[9] = 0xFF, 0xFF, 0xFF, 0xFF

	_, err := DecodeFrame(bytes.NewReader(header))
	if !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("got %v, want ErrPayloadTooLarge", err)
	}
}

func TestEncodeRejectsOversizedPayload(t *testing.T) {
	_, err := Frame{Type: FrameData, Payload: make([]byte, MaxPayloadSize+1)}.Encode()
	if !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("got %v, want ErrPayloadTooLarge", err)
	}
}

func TestUnknownFrameTypeRejected(t *testing.T) {
	t.Run("on encode", func(t *testing.T) {
		if _, err := (Frame{Type: FrameType(99)}).Encode(); !errors.Is(err, ErrUnknownFrame) {
			t.Errorf("got %v, want ErrUnknownFrame", err)
		}
	})

	t.Run("on decode", func(t *testing.T) {
		header := make([]byte, HeaderSize)
		header[0] = 99
		if _, err := DecodeFrame(bytes.NewReader(header)); !errors.Is(err, ErrUnknownFrame) {
			t.Errorf("got %v, want ErrUnknownFrame", err)
		}
	})
}

// An empty payload must decode as nil, not as an empty non-nil slice, so a
// decoded frame equals the frame that produced it.
func TestEmptyPayloadDecodesAsNil(t *testing.T) {
	wire, err := Frame{Type: FrameClose, StreamID: 5}.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := DecodeFrame(bytes.NewReader(wire))
	if err != nil {
		t.Fatalf("DecodeFrame: %v", err)
	}
	if got.Payload != nil {
		t.Errorf("payload = %#v, want nil", got.Payload)
	}
}

func BenchmarkFrameRoundTrip(b *testing.B) {
	frame := Frame{Type: FrameData, StreamID: 1, Payload: make([]byte, 32*1024)}
	b.SetBytes(int64(len(frame.Payload)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		wire, err := frame.Encode()
		if err != nil {
			b.Fatal(err)
		}
		if _, err := DecodeFrame(bytes.NewReader(wire)); err != nil {
			b.Fatal(err)
		}
	}
}
