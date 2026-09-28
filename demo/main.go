package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

func main() {
	port := flag.String("port", "8001", "listen port")
	name := flag.String("name", "", "shown on each response")
	flag.Parse()
	if *name == "" {
		*name = "b" + *port
	}

	var hits atomic.Uint64
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s %d\n", *name, hits.Add(1))
	})

	addr := ":" + *port
	log.Printf("%s %s pid=%d", *name, addr, os.Getpid())
	log.Fatal(http.ListenAndServe(addr, mux))
}
