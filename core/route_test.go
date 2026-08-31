package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestSplitHost(t *testing.T) {
	const base = "pumasi.link"

	valid := []struct {
		host string
		want string
	}{
		{"myapi.pumasi.link", "myapi"},
		{"MyApi.Pumasi.Link", "myapi"},      // Host headers carry any case
		{"myapi.pumasi.link:8080", "myapi"}, // port stripped
		{"myapi.pumasi.link.", "myapi"},     // trailing root dot
		{"  myapi.pumasi.link  ", "myapi"},  // surrounding space
		{"a1-b2.pumasi.link", "a1-b2"},      // digits and inner hyphen
		{"myapi.pumasi.link:443", "myapi"},  // explicit https port
	}
	for _, tc := range valid {
		t.Run(tc.host, func(t *testing.T) {
			got, err := SplitHost(tc.host, base)
			if err != nil {
				t.Fatalf("SplitHost(%q): %v", tc.host, err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

	invalid := []struct {
		name string
		host string
		want error
	}{
		{"empty", "", ErrEmptyHost},
		{"only spaces", "   ", ErrEmptyHost},
		{"apex is not a tunnel", "pumasi.link", ErrForeignHost},
		{"different domain", "myapi.example.com", ErrForeignHost},
		{"suffix lookalike", "evilpumasi.link", ErrForeignHost},
		{"base as prefix", "pumasi.link.evil.com", ErrForeignHost},
		{"empty label", ".pumasi.link", ErrEmptyHost},
		{"nested label", "a.b.pumasi.link", ErrNestedSubname},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			_, err := SplitHost(tc.host, base)
			if !errors.Is(err, tc.want) {
				t.Errorf("SplitHost(%q) = %v, want %v", tc.host, err, tc.want)
			}
		})
	}
}

// A nested host must never resolve to its first label: "a.b.pumasi.link"
// falling through to tunnel "a" would let one tenant serve content under a
// name a visitor reads as belonging to another.
func TestSplitHostRefusesNestedEvenWhenParentRegistered(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if err := r.Register(Tunnel{Subdomain: "myapi", AgentID: "agent-1"}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := r.Lookup("evil.myapi.pumasi.link"); !errors.Is(err, ErrNestedSubname) {
		t.Errorf("got %v, want ErrNestedSubname", err)
	}
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	r := NewRegistry("pumasi.link")
	tunnel := Tunnel{Subdomain: "myapi", AgentID: "agent-1", LocalPort: 3000}
	if err := r.Register(tunnel); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := r.Lookup("myapi.pumasi.link")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.AgentID != "agent-1" || got.LocalPort != 3000 {
		t.Errorf("got %+v, want agent-1/3000", got)
	}
	if url := r.PublicURL("myapi"); url != "https://myapi.pumasi.link" {
		t.Errorf("PublicURL = %q", url)
	}
	if !r.Has("MYAPI") {
		t.Error("Has should be case-insensitive")
	}
	if r.Len() != 1 {
		t.Errorf("Len = %d, want 1", r.Len())
	}
}

func TestRegistryRejectsDuplicateName(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if err := r.Register(Tunnel{Subdomain: "myapi", AgentID: "a"}); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := r.Register(Tunnel{Subdomain: "myapi", AgentID: "b"})
	if !errors.Is(err, ErrNameTaken) {
		t.Fatalf("got %v, want ErrNameTaken", err)
	}
}

// Register takes only already-normalised names, so a mixed-case request is a
// caller error rather than a silent rewrite. Lookup stays case-insensitive
// because a Host header's case is not the caller's choice.
func TestRegistryRegisterRequiresNormalisedName(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if err := r.Register(Tunnel{Subdomain: "MyApi", AgentID: "a"}); !errors.Is(err, ErrSubdomainCharset) {
		t.Fatalf("got %v, want ErrSubdomainCharset", err)
	}
	if r.Len() != 0 {
		t.Error("a rejected Register must leave the registry empty")
	}
}

func TestRegistryRejectsInvalidName(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if err := r.Register(Tunnel{Subdomain: "admin"}); !errors.Is(err, ErrSubdomainReserved) {
		t.Errorf("reserved: got %v", err)
	}
	if err := r.Register(Tunnel{Subdomain: "aa"}); !errors.Is(err, ErrSubdomainTooShort) {
		t.Errorf("too short: got %v", err)
	}
}

func TestRegistryUnknownHost(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if _, err := r.Lookup("nobody.pumasi.link"); !errors.Is(err, ErrNoTunnel) {
		t.Errorf("got %v, want ErrNoTunnel", err)
	}
}

func TestRegistryTCPPorts(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if err := r.Register(Tunnel{Subdomain: "database", AgentID: "a", TCPPort: 20001, LocalPort: 5432}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := r.LookupTCP(20001)
	if err != nil {
		t.Fatalf("LookupTCP: %v", err)
	}
	if got.LocalPort != 5432 {
		t.Errorf("LocalPort = %d, want 5432", got.LocalPort)
	}

	// A second tunnel must not be able to claim a port already serving one.
	err = r.Register(Tunnel{Subdomain: "other", AgentID: "b", TCPPort: 20001})
	if err == nil {
		t.Fatal("expected a conflict on the reused TCP port")
	}
	// The failed registration must not have half-registered the name.
	if r.Has("other") {
		t.Error("failed Register left the subdomain registered")
	}

	if _, err := r.LookupTCP(29999); !errors.Is(err, ErrNoTunnel) {
		t.Errorf("unknown port: got %v, want ErrNoTunnel", err)
	}
}

func TestRegistryUnregisterReleasesNameAndPort(t *testing.T) {
	r := NewRegistry("pumasi.link")
	if err := r.Register(Tunnel{Subdomain: "myapi", AgentID: "a", TCPPort: 20002}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if !r.Unregister("MyApi") {
		t.Fatal("Unregister should report removal and be case-insensitive")
	}
	if r.Unregister("myapi") {
		t.Error("second Unregister should report nothing removed")
	}
	if _, err := r.LookupTCP(20002); !errors.Is(err, ErrNoTunnel) {
		t.Error("TCP port was not released")
	}
	// The freed name must be registrable again, including its port.
	if err := r.Register(Tunnel{Subdomain: "myapi", AgentID: "b", TCPPort: 20002}); err != nil {
		t.Errorf("re-Register after release: %v", err)
	}
}

// A dropped agent connection must free everything it owned — this is the
// path that runs when a laptop closes, and a leak here would slowly consume
// the namespace.
func TestRegistryUnregisterAgent(t *testing.T) {
	r := NewRegistry("pumasi.link")
	for _, name := range []string{"one", "two", "three"} {
		if err := r.Register(Tunnel{Subdomain: name, AgentID: "agent-1"}); err != nil {
			t.Fatalf("Register %s: %v", name, err)
		}
	}
	if err := r.Register(Tunnel{Subdomain: "other", AgentID: "agent-2"}); err != nil {
		t.Fatalf("Register other: %v", err)
	}

	freed := r.UnregisterAgent("agent-1")
	if len(freed) != 3 {
		t.Errorf("freed %v, want 3 subdomains", freed)
	}
	if r.Len() != 1 || !r.Has("other") {
		t.Errorf("agent-2's tunnel should survive; Len = %d", r.Len())
	}
	if got := r.UnregisterAgent("agent-1"); len(got) != 0 {
		t.Errorf("second UnregisterAgent freed %v, want nothing", got)
	}
}

// The relay registers from its control loop while serving requests on many
// goroutines; run with -race to make this meaningful.
func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry("pumasi.link")
	const agents = 16

	var wg sync.WaitGroup
	for i := 0; i < agents; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("agent%03d", i)
			if err := r.Register(Tunnel{Subdomain: name, AgentID: name, TCPPort: 21000 + i}); err != nil {
				t.Errorf("Register %s: %v", name, err)
				return
			}
			if _, err := r.Lookup(name + ".pumasi.link"); err != nil {
				t.Errorf("Lookup %s: %v", name, err)
			}
			r.UnregisterAgent(name)
		}(i)
	}
	wg.Wait()

	if r.Len() != 0 {
		t.Errorf("Len = %d after every agent disconnected, want 0", r.Len())
	}
}

func TestRegistryBaseDomainNormalised(t *testing.T) {
	r := NewRegistry("  Pumasi.Link.  ")
	if r.BaseDomain() != "pumasi.link" {
		t.Fatalf("BaseDomain = %q", r.BaseDomain())
	}
	if err := r.Register(Tunnel{Subdomain: "myapi"}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := r.Lookup("myapi.pumasi.link"); err != nil {
		t.Errorf("Lookup against normalised base: %v", err)
	}
}

func TestValidateSubdomain(t *testing.T) {
	valid := []string{"abc", "myapi", "my-api", "a1b2c3", "x" + strings.Repeat("y", MaxSubdomainLen-1)}
	for _, name := range valid {
		if err := ValidateSubdomain(name); err != nil {
			t.Errorf("ValidateSubdomain(%q) = %v, want nil", name, err)
		}
	}

	invalid := []struct {
		name string
		want error
	}{
		{"", ErrSubdomainTooShort},
		{"ab", ErrSubdomainTooShort},
		{strings.Repeat("a", MaxSubdomainLen+1), ErrSubdomainTooLong},
		{"MyApi", ErrSubdomainCharset},  // uppercase would vary between cert and table
		{"my_api", ErrSubdomainCharset}, // underscore is not a legal DNS label
		{"my.api", ErrSubdomainCharset}, // a dot would smuggle in a nested name
		{"my api", ErrSubdomainCharset},
		{"münchen", ErrSubdomainCharset}, // confusable with an ASCII name
		{"-abc", ErrSubdomainHyphen},
		{"abc-", ErrSubdomainHyphen},
		{"admin", ErrSubdomainReserved},
		{"www", ErrSubdomainReserved},
		{"_acme-challenge", ErrSubdomainCharset}, // underscore rejected before the reserved check
	}
	for _, tc := range invalid {
		if err := ValidateSubdomain(tc.name); !errors.Is(err, tc.want) {
			t.Errorf("ValidateSubdomain(%q) = %v, want %v", tc.name, err, tc.want)
		}
	}
}

// A deterministic source so allocation is testable; it cycles the alphabet.
type seqRand struct {
	next int
	fail error
}

func (s *seqRand) Intn(n int) (int, error) {
	if s.fail != nil {
		return 0, s.fail
	}
	v := s.next % n
	s.next++
	return v, nil
}

func TestGenerateSubdomain(t *testing.T) {
	name, err := GenerateSubdomain(&seqRand{})
	if err != nil {
		t.Fatalf("GenerateSubdomain: %v", err)
	}
	if len(name) != GeneratedSubdomainLen {
		t.Errorf("length = %d, want %d", len(name), GeneratedSubdomainLen)
	}
	// Whatever the source yields, the result must be a legal subdomain.
	if err := ValidateSubdomain(name); err != nil {
		t.Errorf("generated %q is not valid: %v", name, err)
	}
}

func TestGenerateSubdomainRealRandomnessIsValid(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		name, err := GenerateSubdomain(nil) // crypto/rand
		if err != nil {
			t.Fatalf("GenerateSubdomain: %v", err)
		}
		if err := ValidateSubdomain(name); err != nil {
			t.Fatalf("generated %q is not valid: %v", name, err)
		}
		seen[name] = true
	}
	// 200 draws from a ~50-bit space should never repeat; a collision here
	// means the source is not actually random.
	if len(seen) != 200 {
		t.Errorf("got %d distinct names from 200 draws", len(seen))
	}
}

func TestAllocateSubdomainSkipsTakenNames(t *testing.T) {
	r := NewRegistry("pumasi.link")
	source := &seqRand{}

	first, err := AllocateSubdomain(r, source)
	if err != nil {
		t.Fatalf("AllocateSubdomain: %v", err)
	}
	if err := r.Register(Tunnel{Subdomain: first, AgentID: "a"}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Rewind the source so it would produce the same name again; allocation
	// must notice the collision and move on.
	source.next = 0
	second, err := AllocateSubdomain(r, source)
	if err != nil {
		t.Fatalf("second AllocateSubdomain: %v", err)
	}
	if second == first {
		t.Errorf("allocated the taken name %q twice", first)
	}
}

func TestAllocateSubdomainPropagatesSourceFailure(t *testing.T) {
	wantErr := errors.New("entropy pool drained")
	_, err := AllocateSubdomain(nil, &seqRand{fail: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v, want the source's error", err)
	}
}

func BenchmarkRegistryLookup(b *testing.B) {
	r := NewRegistry("pumasi.link")
	for i := 0; i < 1000; i++ {
		if err := r.Register(Tunnel{Subdomain: fmt.Sprintf("tunnel%04d", i), AgentID: "a"}); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Lookup("tunnel0500.pumasi.link"); err != nil {
			b.Fatal(err)
		}
	}
}
