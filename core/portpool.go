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
type PortPool struct {
	low, high int

	mu       sync.Mutex
	inUse    map[int]string // port -> owning subdomain
	cursor   int
	reserved map[int]bool
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
	}
	for _, port := range reserved {
		p.reserved[port] = true
	}
	return p, nil
}

// Allocate reserves the next free port for owner and returns it.
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
	p.inUse[port] = owner
	return nil
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
