package core

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrPortPoolExhausted = errors.New("core: no free public TCP port in the pool")
	ErrPortOutOfRange    = errors.New("core: port is outside the pool range")
	ErrPortNotAllocated  = errors.New("core: port is not allocated")
	ErrBadPortRange      = errors.New("core: invalid port range")
)

// PortPool hands out public TCP ports for raw (non-HTTP) tunnels. Allocation
// walks forward from the last handed-out port and wraps, so a released port
// is not immediately reissued: a client that reconnects seconds after a crash
// should not inherit a port that another tenant's peers are still dialling.
//
// The pool is safe for concurrent use and does no I/O — it decides which
// number to use; binding the listener is the relay's job.
//
// A number is in one of three states, and the third is why an address can be
// written down (spec/0004 §5.2):
//
//	reserved — the operator keeps it clear. Never allocated to anyone, ever.
//	inUse    — allocated to a live tunnel right now.
//	held     — claimed by a tenant, not currently allocated.
//
// Without the third, a tenant's port is free the instant its agent drops and
// the next anonymous request walks straight onto it — which is the same defect
// as a name with no owner, one address space over.
type PortPool struct {
	low, high int

	mu       sync.Mutex
	inUse    map[int]string // port -> owning tenant
	cursor   int
	reserved map[int]bool
	held     map[int]string // port -> tenant it belongs to between connections
}

// NewPortPool returns a pool over the inclusive range [low, high]. Ports the
// operator wants kept clear (an ssh daemon, a metrics endpoint) are passed as
// reserved and never allocated.
func NewPortPool(low, high int, reserved ...int) (*PortPool, error) {
	if low < 1 || high > 65535 || low > high {
		return nil, fmt.Errorf("%w: [%d, %d]", ErrBadPortRange, low, high)
	}
	p := &PortPool{
		low:      low,
		high:     high,
		inUse:    make(map[int]string),
		cursor:   low,
		reserved: make(map[int]bool, len(reserved)),
		held:     make(map[int]string),
	}
	for _, port := range reserved {
		p.reserved[port] = true
	}
	return p, nil
}

// Allocate reserves the next free port for owner and returns it.
//
// A port held by a different tenant is skipped rather than refused: this
// caller asked for no particular number, so walking past one costs it nothing
// and costs the holder its address otherwise.
func (p *PortPool) Allocate(owner string) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	span := p.high - p.low + 1
	for i := 0; i < span; i++ {
		port := p.cursor
		p.cursor++
		if p.cursor > p.high {
			p.cursor = p.low
		}
		if p.reserved[port] || p.inUse[port] != "" {
			continue
		}
		if h := p.held[port]; h != "" && h != owner {
			continue
		}
		p.inUse[port] = owner
		return port, nil
	}
	return 0, fmt.Errorf("%w: [%d, %d] all in use", ErrPortPoolExhausted, p.low, p.high)
}

// AllocateSpecific reserves an exact port, for a tenant whose port must
// survive reconnects. It fails if the port is out of range, reserved, or
// already held.
func (p *PortPool) AllocateSpecific(port int, owner string) error {
	if port < p.low || port > p.high {
		return fmt.Errorf("%w: %d not in [%d, %d]", ErrPortOutOfRange, port, p.low, p.high)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.reserved[port] {
		return fmt.Errorf("%w: %d is reserved by the operator", ErrPortOutOfRange, port)
	}
	if holder := p.inUse[port]; holder != "" {
		return fmt.Errorf("core: port %d is already held by %q", port, holder)
	}
	// A held port is its tenant's to reclaim and nobody else's to ask for.
	// This is what makes a --tcp-port survive the seconds an agent spends
	// reconnecting, rather than surviving them by luck.
	if h := p.held[port]; h != "" && h != owner {
		return fmt.Errorf("%w: %d is reserved by %q", ErrPortOutOfRange, port, h)
	}
	p.inUse[port] = owner
	return nil
}

// Hold marks a port as belonging to a tenant even while nothing is allocated
// on it, so it is neither walked onto by Allocate nor handed to a stranger
// that names the number. Holding a port already held by someone else is an
// error rather than a takeover.
//
// A hold is bookkeeping, not authority: core.Reservations decides who owns
// what, and the relay is where the two meet.
func (p *PortPool) Hold(port int, holder string) error {
	if port < p.low || port > p.high {
		return fmt.Errorf("%w: %d not in [%d, %d]", ErrPortOutOfRange, port, p.low, p.high)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.reserved[port] {
		return fmt.Errorf("%w: %d is reserved by the operator", ErrPortOutOfRange, port)
	}
	if h := p.held[port]; h != "" && h != holder {
		return fmt.Errorf("%w: %d is reserved by %q", ErrPortOutOfRange, port, h)
	}
	p.held[port] = holder
	return nil
}

// Unhold gives up a hold. Releasing a hold nobody has is a no-op, because the
// caller's intent — nobody holds this — is already true.
func (p *PortPool) Unhold(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.held, port)
}

// Holder reports which tenant a port belongs to between connections, or "".
func (p *PortPool) Holder(port int) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.held[port]
}

// Release returns a port to the pool. Releasing a port that is not allocated
// is an error rather than a silent no-op, because it means the caller's
// bookkeeping disagrees with the pool's.
func (p *PortPool) Release(port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.inUse[port] == "" {
		return fmt.Errorf("%w: %d", ErrPortNotAllocated, port)
	}
	delete(p.inUse, port)
	return nil
}

// ReleaseOwner frees every port held by an owner and reports them, which is
// what the relay calls when an agent disconnects.
func (p *PortPool) ReleaseOwner(owner string) []int {
	p.mu.Lock()
	defer p.mu.Unlock()
	var freed []int
	for port, holder := range p.inUse {
		if holder == owner {
			delete(p.inUse, port)
			freed = append(freed, port)
		}
	}
	return freed
}

// Owner reports which tenant holds a port, or "" if it is free.
func (p *PortPool) Owner(port int) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.inUse[port]
}

// InUse reports how many ports are currently allocated.
func (p *PortPool) InUse() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.inUse)
}

// Capacity reports how many ports the pool can ever hand out.
func (p *PortPool) Capacity() int {
	return p.high - p.low + 1 - len(p.reserved)
}
