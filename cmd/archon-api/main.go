package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hacrex/Archon-Base/internal/server"
)

func main() {
	addr := os.Getenv("ARCHON_API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := server.New()
	log.Printf("archon-api listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
