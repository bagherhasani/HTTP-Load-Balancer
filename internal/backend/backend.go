package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy

	mu      sync.RWMutex
	healthy bool
}

func New(raw string) (*Backend, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &Backend{
		URL:     u,
		Proxy:   httputil.NewSingleHostReverseProxy(u),
		healthy: true,
	}, nil
}

func (b *Backend) Healthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.healthy
}

func (b *Backend) SetHealthy(ok bool) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	changed := b.healthy != ok
	b.healthy = ok
	return changed
}
