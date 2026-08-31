package core

import (
	"errors"
	"sync"
	"testing"
)

func TestNewPortPoolValidatesRange(t *testing.T) {
	bad := [][2]int{{0, 100}, {100, 99}, {1, 70000}, {-1, 10}}
	for _, r := range bad {
		if _, err := NewPortPool(r[0], r[1]); !errors.Is(err, ErrBadPortRange) {
			t.Errorf("NewPortPool(%d, %d) = %v, want ErrBadPortRange", r[0], r[1], err)
		}
	}
	if _, err := NewPortPool(20000, 20000); err != nil {
		t.Errorf("single-port range should be legal: %v", err)
	}
}

func TestPortPoolAllocateAndRelease(t *testing.T) {
	p, err := NewPortPool(20000, 20002)
	if err != nil {
		t.Fatalf("NewPortPool: %v", err)
	}
	if p.Capacity() != 3 {
		t.Errorf("Capacity = %d, want 3", p.Capacity())
	}

	seen := map[int]bool{}
	for i := 0; i < 3; i++ {
		port, err := p.Allocate("tenant")
		if err != nil {
			t.Fatalf("Allocate %d: %v", i, err)
		}
		if port < 20000 || port > 20002 {
			t.Fatalf("port %d outside the range", port)
		}
		if seen[port] {
			t.Fatalf("port %d handed out twice", port)
		}
		seen[port] = true
	}
	if p.InUse() != 3 {
		t.Errorf("InUse = %d, want 3", p.InUse())
	}

	if _, err := p.Allocate("tenant"); !errors.Is(err, ErrPortPoolExhausted) {
		t.Errorf("exhausted pool: got %v, want ErrPortPoolExhausted", err)
	}

	if err := p.Release(20001); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if got := p.Owner(20001); got != "" {
		t.Errorf("Owner after release = %q, want empty", got)
	}
	if port, err := p.Allocate("next"); err != nil || port != 20001 {
		t.Errorf("Allocate after release = %d, %v; want 20001", port, err)
	}
}

// Releasing something never allocated means the caller's bookkeeping has
// drifted from the pool's; that must surface rather than pass quietly.
func TestPortPoolReleaseUnallocated(t *testing.T) {
	p, _ := NewPortPool(20000, 20010)
	if err := p.Release(20005); !errors.Is(err, ErrPortNotAllocated) {
		t.Errorf("got %v, want ErrPortNotAllocated", err)
	}
}

// A reconnecting client must not immediately inherit a just-released port,
// because peers may still be dialling it for the previous tenant.
func TestPortPoolDoesNotImmediatelyReuseReleasedPort(t *testing.T) {
	p, _ := NewPortPool(20000, 20009)

	first, err := p.Allocate("tenant-a")
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if err := p.Release(first); err != nil {
		t.Fatalf("Release: %v", err)
	}

	second, err := p.Allocate("tenant-b")
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if second == first {
		t.Errorf("port %d was reissued immediately after release", first)
	}
}

func TestPortPoolReservedPortsNeverAllocated(t *testing.T) {
	p, err := NewPortPool(20000, 20003, 20001, 20002)
	if err != nil {
		t.Fatalf("NewPortPool: %v", err)
	}
	if p.Capacity() != 2 {
		t.Errorf("Capacity = %d, want 2 (four ports minus two reserved)", p.Capacity())
	}

	for i := 0; i < 2; i++ {
		port, err := p.Allocate("tenant")
		if err != nil {
			t.Fatalf("Allocate: %v", err)
		}
		if port == 20001 || port == 20002 {
			t.Fatalf("allocated reserved port %d", port)
		}
	}
	if _, err := p.Allocate("tenant"); !errors.Is(err, ErrPortPoolExhausted) {
		t.Errorf("got %v, want ErrPortPoolExhausted", err)
	}

	if err := p.AllocateSpecific(20001, "tenant"); !errors.Is(err, ErrPortOutOfRange) {
		t.Errorf("AllocateSpecific on a reserved port: got %v", err)
	}
}

func TestPortPoolAllocateSpecific(t *testing.T) {
	p, _ := NewPortPool(20000, 20010)

	if err := p.AllocateSpecific(20005, "reserved-tenant"); err != nil {
		t.Fatalf("AllocateSpecific: %v", err)
	}
	if got := p.Owner(20005); got != "reserved-tenant" {
		t.Errorf("Owner = %q", got)
	}

	if err := p.AllocateSpecific(20005, "other"); err == nil {
		t.Error("expected a conflict claiming a held port")
	}
	for _, port := range []int{19999, 20011} {
		if err := p.AllocateSpecific(port, "x"); !errors.Is(err, ErrPortOutOfRange) {
			t.Errorf("AllocateSpecific(%d) = %v, want ErrPortOutOfRange", port, err)
		}
	}

	// A specifically-held port must not also be handed out by Allocate.
	for i := 0; i < p.Capacity()-1; i++ {
		if port, err := p.Allocate("bulk"); err != nil {
			t.Fatalf("Allocate: %v", err)
		} else if port == 20005 {
			t.Fatal("Allocate handed out the specifically-held port")
		}
	}
}

func TestPortPoolReleaseOwner(t *testing.T) {
	p, _ := NewPortPool(20000, 20010)
	for i := 0; i < 3; i++ {
		if _, err := p.Allocate("agent-1"); err != nil {
			t.Fatalf("Allocate: %v", err)
		}
	}
	if _, err := p.Allocate("agent-2"); err != nil {
		t.Fatalf("Allocate: %v", err)
	}

	freed := p.ReleaseOwner("agent-1")
	if len(freed) != 3 {
		t.Errorf("freed %v, want 3 ports", freed)
	}
	if p.InUse() != 1 {
		t.Errorf("InUse = %d, want 1 (agent-2 keeps its port)", p.InUse())
	}
	if got := p.ReleaseOwner("agent-1"); len(got) != 0 {
		t.Errorf("second ReleaseOwner freed %v, want nothing", got)
	}
}

// Two agents connecting at once must never receive the same port; run with
// -race for this to mean anything.
func TestPortPoolConcurrentAllocation(t *testing.T) {
	p, _ := NewPortPool(20000, 20099)
	const workers = 50

	var (
		mu     sync.Mutex
		ports  = map[int]bool{}
		wg     sync.WaitGroup
		errsMu sync.Mutex
		errs   []error
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			port, err := p.Allocate("tenant")
			if err != nil {
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if ports[port] {
				t.Errorf("port %d allocated twice", port)
			}
			ports[port] = true
		}()
	}
	wg.Wait()

	if len(errs) != 0 {
		t.Fatalf("allocation errors: %v", errs)
	}
	if len(ports) != workers {
		t.Errorf("got %d distinct ports from %d workers", len(ports), workers)
	}
}
