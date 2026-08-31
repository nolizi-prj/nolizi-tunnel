package relay

import (
	"net"

	"github.com/pumasi-ai/pumasi-tunnel/core"
	"github.com/pumasi-ai/pumasi-tunnel/mux"
)

// tunnelSession is a live connection back to whoever is publishing a port.
// The relay talks to this and never cares how the bytes get home, which is
// what lets one routing table serve two very different clients:
//
//   - the native agent, multiplexing streams over our own frame protocol;
//   - a stock ssh client doing `ssh -R`, where each visitor becomes an SSH
//     channel and there is no binary to install at all.
//
// The second is the reason this abstraction exists. The incumbent study found
// zero-install SSH to be the highest-leverage onboarding behaviour in the
// category (docs/ux/incumbent-ux-spec.md §1, checklist item 4), and it cannot
// be bolted on later without a seam like this.
type tunnelSession interface {
	// OpenStream returns a new connection to the publisher, for one visitor.
	OpenStream(tcp bool) (net.Conn, error)
	// CloseChan closes when the session ends, however it ended.
	CloseChan() <-chan struct{}
	// Close tears the session down.
	Close() error
	// Kind names the client for logs and the console: "agent" or "ssh".
	Kind() string
}

// muxSession adapts the native agent's multiplexed connection.
type muxSession struct{ s *mux.Session }

func (m muxSession) OpenStream(tcp bool) (net.Conn, error) {
	var flags uint8
	if tcp {
		flags = core.FlagTCP
	}
	stream, err := m.s.Open(flags)
	if err != nil {
		return nil, err
	}
	return stream.NetConn(), nil
}

func (m muxSession) CloseChan() <-chan struct{} { return m.s.CloseChan() }
func (m muxSession) Close() error               { return m.s.Close() }
func (m muxSession) Kind() string               { return "agent" }
