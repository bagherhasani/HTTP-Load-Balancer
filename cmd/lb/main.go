package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bagherhasani/HTTP-Load-Balancer/internal/backend"
	"github.com/bagherhasani/HTTP-Load-Balancer/internal/pool"
	"gopkg.in/yaml.v3"
)

type cfg struct {
	Listen   string   `yaml:"listen"`
	Backends []string `yaml:"backends"`
	Health   struct {
		Path  string `yaml:"path"`
		Every string `yaml:"every"`
	} `yaml:"health"`
}

func main() {
	path := flag.String("config", "configs/config.yaml", "config file")
	flag.Parse()

	c, err := load(*path)
	if err != nil {
		log.Fatal(err)
	}

	var backends []*backend.Backend
	for _, raw := range c.Backends {
		b, err := backend.New(raw)
		if err != nil {
			log.Fatalf("backend %q: %v", raw, err)
		}
		b.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("%s: %v", b.URL.Host, err)
			if b.SetHealthy(false) {
				log.Printf("%s marked down", b.URL.Host)
			}
			http.Error(w, "backend error", http.StatusBadGateway)
		}
		backends = append(backends, b)
	}
	if len(backends) == 0 {
		log.Fatal("no backends in config")
	}

	p := pool.New(backends)
	interval := 5 * time.Second
	if c.Health.Every != "" {
		interval, err = time.ParseDuration(c.Health.Every)
		if err != nil {
			log.Fatalf("health.every: %v", err)
		}
	}
	healthPath := c.Health.Path
	if healthPath == "" {
		healthPath = "/health"
	}

	go watch(p, healthPath, interval)

	listen := c.Listen
	if listen == "" {
		listen = ":8080"
	}
	log.Printf("listening on %s, %d backends, health %s every %s", listen, len(backends), healthPath, interval)
	log.Fatal(http.ListenAndServe(listen, &handler{pool: p}))
}

type handler struct {
	pool *pool.Pool
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b := h.pool.Next()
	if b == nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
		return
	}
	b.Proxy.ServeHTTP(w, r)
}

func watch(p *pool.Pool, path string, every time.Duration) {
	client := &http.Client{Timeout: 2 * time.Second}
	tick := time.NewTicker(every)
	defer tick.Stop()

	check := func() {
		for _, b := range p.All() {
			u := *b.URL
			u.Path = path
			ok := false
			resp, err := client.Get(u.String())
			if err == nil {
				resp.Body.Close()
				ok = resp.StatusCode == http.StatusOK
			}
			if b.SetHealthy(ok) {
				log.Printf("%s healthy=%v", b.URL.Host, ok)
			}
		}
	}

	check()
	for range tick.C {
		check()
	}
}

func load(path string) (cfg, error) {
	var c cfg
	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}
