// Package core is the pure core of Pumasi Tunnel: the wire protocol, the
// host-to-tunnel routing table, subdomain allocation, and the public TCP port
// pool. Nothing here opens a socket, reads a clock, or touches a database —
// every function is deterministic in its inputs, so the whole protocol layer
// is testable without a network. The relay and the CLI are thin I/O shells
// around this package.
package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// FrameType identifies what a frame carries. Values are wire constants and
// must never be renumbered — an old client and a new relay have to agree.
type FrameType uint8

const (
	FrameOpen   FrameType = 1 // client or relay opens a virtual stream
	FrameData   FrameType = 2 // payload bytes for an open stream
	FrameClose  FrameType = 3 // one side is done writing to a stream
	FramePing   FrameType = 4 // liveness probe; payload echoed back
	FramePong   FrameType = 5 // response to Ping, same payload
	FrameAuth   FrameType = 6 // client presents its token and requested name
	FrameAuthOK FrameType = 7 // relay accepts, payload carries the public URL
	FrameError  FrameType = 8 // relay refuses; payload is a human-readable reason
)

// String makes frame types readable in test failures and logs.
func (t FrameType) String() string {
	switch t {
	case FrameOpen:
		return "OPEN"
	case FrameData:
		return "DATA"
	case FrameClose:
		return "CLOSE"
	case FramePing:
		return "PING"
	case FramePong:
		return "PONG"
	case FrameAuth:
		return "AUTH"
	case FrameAuthOK:
		return "AUTH_OK"
	case FrameError:
		return "ERROR"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", uint8(t))
	}
}

// Frame flags. Flags are a bitfield so a frame can carry several at once.
const (
	// FlagEndStream marks the last DATA frame of a stream, letting a reader
	// finish without waiting for a separate CLOSE.
	FlagEndStream uint8 = 1 << 0
	// FlagTCP marks an OPEN as raw TCP rather than HTTP, so the relay skips
	// host-header parsing and copies bytes verbatim.
	FlagTCP uint8 = 1 << 1
)

// HeaderSize is the fixed byte length of a frame header: type, flags,
// stream id, payload length.
const HeaderSize = 1 + 1 + 4 + 4

// MaxPayloadSize caps a single frame's payload. A relay must be able to
// reject a hostile length prefix before allocating, so this bound is part of
// the protocol rather than a local policy.
const MaxPayloadSize = 1 << 20 // 1 MiB

// Protocol errors. Callers match on these rather than on message text.
var (
	ErrPayloadTooLarge = errors.New("core: frame payload exceeds MaxPayloadSize")
	ErrUnknownFrame    = errors.New("core: unknown frame type")
	ErrShortHeader     = errors.New("core: truncated frame header")
)

// Frame is one unit of the tunnel protocol. Payload is nil for frames that
// carry no bytes; it is never a zero-length non-nil slice after a decode, so
// round-tripping a frame compares equal with reflect.DeepEqual.
type Frame struct {
	Type     FrameType
	Flags    uint8
	StreamID uint32
	Payload  []byte
}

// Encode serialises a frame to its wire representation.
//
// Layout, all integers big-endian:
//
//	 0        1        2                     6                     10
//	+--------+--------+---------------------+---------------------+
//	| Type   | Flags  |      StreamID       |    Payload length   | payload…
//	+--------+--------+---------------------+---------------------+
func (f Frame) Encode() ([]byte, error) {
	if len(f.Payload) > MaxPayloadSize {
		return nil, fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(f.Payload))
	}
	if !f.Type.valid() {
		return nil, fmt.Errorf("%w: %d", ErrUnknownFrame, uint8(f.Type))
	}
	buf := make([]byte, HeaderSize+len(f.Payload))
	buf[0] = uint8(f.Type)
	buf[1] = f.Flags
	binary.BigEndian.PutUint32(buf[2:6], f.StreamID)
	binary.BigEndian.PutUint32(buf[6:10], uint32(len(f.Payload)))
	copy(buf[HeaderSize:], f.Payload)
	return buf, nil
}

func (t FrameType) valid() bool {
	return t >= FrameOpen && t <= FrameError
}

// DecodeFrame reads exactly one frame from r. It returns io.EOF only when the
// reader was positioned cleanly between frames; a partial header or a partial
// payload is io.ErrUnexpectedEOF, so a caller can tell a graceful close from a
// severed connection.
func DecodeFrame(r io.Reader) (Frame, error) {
	var header [HeaderSize]byte
	n, err := io.ReadFull(r, header[:])
	if err != nil {
		if err == io.EOF && n == 0 {
			return Frame{}, io.EOF
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return Frame{}, fmt.Errorf("%w: got %d of %d bytes", ErrShortHeader, n, HeaderSize)
		}
		return Frame{}, err
	}

	f := Frame{
		Type:     FrameType(header[0]),
		Flags:    header[1],
		StreamID: binary.BigEndian.Uint32(header[2:6]),
	}
	if !f.Type.valid() {
		return Frame{}, fmt.Errorf("%w: %d", ErrUnknownFrame, header[0])
	}

	length := binary.BigEndian.Uint32(header[6:10])
	// Checked before allocating: a forged length prefix must not let a peer
	// make the relay reserve a gigabyte.
	if length > MaxPayloadSize {
		return Frame{}, fmt.Errorf("%w: header declares %d bytes", ErrPayloadTooLarge, length)
	}
	if length == 0 {
		return f, nil
	}

	f.Payload = make([]byte, length)
	if _, err := io.ReadFull(r, f.Payload); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return Frame{}, fmt.Errorf("core: truncated payload: %w", err)
	}
	return f, nil
}
