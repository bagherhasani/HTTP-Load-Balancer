package pool

import (
	"testing"

	"github.com/bagherhasani/HTTP-Load-Balancer/internal/backend"
)

func must(t *testing.T, raw string) *backend.Backend {
	t.Helper()
	b, err := backend.New(raw)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNextSkipsUnhealthy(t *testing.T) {
	a := must(t, "http://127.0.0.1:8001")
	b := must(t, "http://127.0.0.1:8002")
	c := must(t, "http://127.0.0.1:8003")
	b.SetHealthy(false)

	p := New([]*backend.Backend{a, b, c})
	seen := map[string]int{}
	for i := 0; i < 6; i++ {
		got := p.Next()
		if got == nil {
			t.Fatal("pool returned nothing")
		}
		if got == b {
			t.Fatal("picked the down backend")
		}
		seen[got.URL.Host]++
	}
	if seen["127.0.0.1:8001"] == 0 || seen["127.0.0.1:8003"] == 0 {
		t.Fatalf("wanted 8001 and 8003, got %v", seen)
	}
}

func TestNextNilWhenAllDown(t *testing.T) {
	a := must(t, "http://127.0.0.1:8001")
	a.SetHealthy(false)
	p := New([]*backend.Backend{a})
	if p.Next() != nil {
		t.Fatal("expected nil")
	}
}
