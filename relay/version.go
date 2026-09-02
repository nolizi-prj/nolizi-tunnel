package relay

import "github.com/pumasi-ai/pumasi-tunnel/internal/buildinfo"

// Version remains exported for API consumers while binaries use the small
// buildinfo package and avoid pulling the embedded console into the agent.
const Version = buildinfo.Version
