package pool

import (
	"sync/atomic"

	"github.com/bagherhasani/HTTP-Load-Balancer/internal/backend"
)

type Pool struct {
	backends []*backend.Backend
	cursor   atomic.Uint64
}

func New(backends []*backend.Backend) *Pool {
	return &Pool{backends: backends}
}

func (p *Pool) Next() *backend.Backend {
	n := len(p.backends)
	if n == 0 {
		return nil
	}
	start := int(p.cursor.Add(1) - 1)
	for i := 0; i < n; i++ {
		b := p.backends[(start+i)%n]
		if b.Healthy() {
			return b
		}
	}
	return nil
}

func (p *Pool) All() []*backend.Backend {
	return p.backends
}
