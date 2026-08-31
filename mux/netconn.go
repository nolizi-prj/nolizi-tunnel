package mux

import (
	"errors"
	"net"
	"time"
)

// streamAddr names one end of a stream for net.Conn's Local/RemoteAddr. The
// addresses are informational — they appear in logs and in net/http's error
// messages, and nothing dials them.
type streamAddr struct{ label string }

func (a streamAddr) Network() string { return "pumasi-mux" }
func (a streamAddr) String() string  { return a.label }

// ErrDeadlineUnsupported is returned by the deadline setters. Streams inherit
// the lifetime of their session and are cancelled by closing them, so there
// is no per-stream timer to set.
var ErrDeadlineUnsupported = errors.New("mux: stream deadlines are not supported; close the stream to cancel")

// NetConn presents a stream as a net.Conn so it can be handed to net/http,
// crypto/tls, or anything else that expects one. Read, Write and Close are
// the stream's own; only the addressing and deadline methods are added.
//
// The deadline methods report ErrDeadlineUnsupported rather than pretending
// to succeed. net/http tolerates that — it falls back to context
// cancellation, which closes the stream — and a silent no-op would let a
// caller believe a timeout was armed when nothing would ever fire it.
func (st *Stream) NetConn() net.Conn { return &streamConn{Stream: st} }

type streamConn struct{ *Stream }

func (c *streamConn) LocalAddr() net.Addr {
	return streamAddr{label: "stream/local"}
}

func (c *streamConn) RemoteAddr() net.Addr {
	return streamAddr{label: "stream/remote"}
}

func (c *streamConn) SetDeadline(time.Time) error      { return ErrDeadlineUnsupported }
func (c *streamConn) SetReadDeadline(time.Time) error  { return ErrDeadlineUnsupported }
func (c *streamConn) SetWriteDeadline(time.Time) error { return ErrDeadlineUnsupported }

// Compile-time proof that the adapter really satisfies net.Conn.
var _ net.Conn = (*streamConn)(nil)
