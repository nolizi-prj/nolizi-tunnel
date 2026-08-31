package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// The handshake happens once, on the raw connection, before either side wraps
// it in a multiplexer: the agent sends AuthRequest in a FrameAuth, and the
// relay answers with AuthResponse in a FrameAuthOK or a reason in a
// FrameError. Doing it before the mux keeps the multiplexer free of any
// notion of identity, and keeps this decision — who gets which name — pure
// and testable here.

// AuthRequest is what an agent asks for when it connects.
type AuthRequest struct {
	// Token authenticates a reserved subdomain. Empty means anonymous, which
	// the relay may answer with an ephemeral name.
	Token string `json:"token,omitempty"`
	// Subdomain is the name the agent wants. Empty asks the relay to choose.
	Subdomain string `json:"subdomain,omitempty"`
	// LocalPort is what the agent forwards to, recorded for display.
	LocalPort int `json:"local_port,omitempty"`
	// TCP asks for a public TCP port instead of an HTTP hostname.
	TCP bool `json:"tcp,omitempty"`
	// ClientVersion is reported in the relay's logs and status surface.
	ClientVersion string `json:"client_version,omitempty"`
}

// AuthResponse is the relay's answer to an accepted agent.
type AuthResponse struct {
	Subdomain string `json:"subdomain"`
	URL       string `json:"url"`
	// TCPAddr is set for raw TCP tunnels, as "host:port".
	TCPAddr string `json:"tcp_addr,omitempty"`
	// AgentID identifies this session in relay logs.
	AgentID string `json:"agent_id"`
}

// MaxHandshakeBytes bounds a handshake payload. Generous for the struct above
// and far below MaxPayloadSize, so a peer cannot make the relay parse
// megabytes of JSON before it has authenticated anything.
const MaxHandshakeBytes = 4096

// EncodeAuthRequest builds the frame an agent sends first.
func EncodeAuthRequest(req AuthRequest) (Frame, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return Frame{}, fmt.Errorf("core: encoding auth request: %w", err)
	}
	if len(payload) > MaxHandshakeBytes {
		return Frame{}, fmt.Errorf("%w: auth request is %d bytes", ErrPayloadTooLarge, len(payload))
	}
	return Frame{Type: FrameAuth, Payload: payload}, nil
}

// DecodeAuthRequest reads an agent's opening frame.
func DecodeAuthRequest(f Frame) (AuthRequest, error) {
	if f.Type != FrameAuth {
		return AuthRequest{}, fmt.Errorf("core: expected AUTH, got %s", f.Type)
	}
	if len(f.Payload) > MaxHandshakeBytes {
		return AuthRequest{}, fmt.Errorf("%w: auth request is %d bytes", ErrPayloadTooLarge, len(f.Payload))
	}
	var req AuthRequest
	if err := json.Unmarshal(f.Payload, &req); err != nil {
		return AuthRequest{}, fmt.Errorf("core: decoding auth request: %w", err)
	}
	req.Subdomain = strings.ToLower(strings.TrimSpace(req.Subdomain))
	return req, nil
}

// EncodeAuthResponse builds the relay's acceptance frame.
func EncodeAuthResponse(resp AuthResponse) (Frame, error) {
	payload, err := json.Marshal(resp)
	if err != nil {
		return Frame{}, fmt.Errorf("core: encoding auth response: %w", err)
	}
	return Frame{Type: FrameAuthOK, Payload: payload}, nil
}

// DecodeAuthResponse reads the relay's answer. A FrameError is returned as an
// error carrying the relay's stated reason, so the CLI can print exactly what
// the relay refused and why.
func DecodeAuthResponse(f Frame) (AuthResponse, error) {
	switch f.Type {
	case FrameAuthOK:
		var resp AuthResponse
		if err := json.Unmarshal(f.Payload, &resp); err != nil {
			return AuthResponse{}, fmt.Errorf("core: decoding auth response: %w", err)
		}
		return resp, nil
	case FrameError:
		return AuthResponse{}, fmt.Errorf("relay refused the tunnel: %s", strings.TrimSpace(string(f.Payload)))
	default:
		return AuthResponse{}, fmt.Errorf("core: expected AUTH_OK or ERROR, got %s", f.Type)
	}
}

// ErrorFrame builds the relay's refusal.
func ErrorFrame(reason string) Frame {
	if len(reason) > MaxHandshakeBytes {
		reason = reason[:MaxHandshakeBytes]
	}
	return Frame{Type: FrameError, Payload: []byte(reason)}
}
