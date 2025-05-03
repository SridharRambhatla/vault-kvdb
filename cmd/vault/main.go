package main

import (
	"flag"
	"log"

	"go-kvdb/internal/server"
)

func main() {
	// Parse command line flags
	httpAddr := flag.String("http-addr", "127.0.0.1:8080", "HTTP server address")
	cacheSize := flag.Int64("cache-size", 1024*1024*1024, "Cache size in bytes (default: 1GB)")
	flag.Parse()

	// Create and start server
	srv := server.NewServer(*cacheSize)
	log.Printf("Starting Vault server on %s with cache size %d bytes", *httpAddr, *cacheSize)
	if err := srv.Start(*httpAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
