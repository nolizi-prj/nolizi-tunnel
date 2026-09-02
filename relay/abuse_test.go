package relay

import (
	"testing"
	"time"
)

func TestTunnelAbuseGuardLimitsActiveAndStartRate(t *testing.T) {
	g := newTunnelAbuseGuard(1, 2, time.Minute)
	now := time.Unix(1000, 0)
	release, err := g.acquire("192.0.2.10:1000", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.acquire("192.0.2.10:2000", now); err == nil {
		t.Fatal("second active tunnel from one address was accepted")
	}
	release()
	if _, err := g.acquire("192.0.2.10:3000", now); err == nil {
		t.Fatal("third start inside the rate window was accepted")
	}
	if releaseLater, err := g.acquire("192.0.2.10:4000", now.Add(time.Minute+time.Second)); err != nil {
		t.Fatalf("source remained limited after the window: %v", err)
	} else {
		releaseLater()
	}
}

func TestSourceHostGroupsPortsButNotAddresses(t *testing.T) {
	if got := sourceHost("192.0.2.10:1234"); got != "192.0.2.10" {
		t.Fatalf("sourceHost = %q", got)
	}
	if got := sourceHost("[2001:db8::1]:1234"); got != "2001:db8::1" {
		t.Fatalf("IPv6 sourceHost = %q", got)
	}
}

func TestVisitorLimitIsPerTunnelAndReleases(t *testing.T) {
	r, err := New(Config{BaseDomain: "pumasi.link", MaxConnectionsPerTunnel: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if !r.acquireVisitor("agent-a") {
		t.Fatal("first visitor was refused")
	}
	if r.acquireVisitor("agent-a") {
		t.Fatal("second visitor exceeded the tunnel limit")
	}
	if !r.acquireVisitor("agent-b") {
		t.Fatal("one tunnel consumed another tunnel's allowance")
	}
	r.releaseVisitor("agent-a")
	if !r.acquireVisitor("agent-a") {
		t.Fatal("released visitor slot was not reusable")
	}
}
