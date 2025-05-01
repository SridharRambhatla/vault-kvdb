package main

import (
	"flag"
	"go-kvdb/config"
	"go-kvdb/storage"
	"go-kvdb/web"
	"log"
	"net/http"
)

var (
	dbLocation = flag.String("db-location", "", "The path to the bolt db database")
	httpAddr   = flag.String("http-addr", "127.0.0.1:8080", "HTTP host and port")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
	shard      = flag.String("shard", "", "The name of the shard for data")
	cacheSize  = flag.Int64("cache-size", 1024*1024*1024, "Cache size in bytes (default: 1GB)")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatal("Must pass db-location")
	}

	if *shard == "" {
		log.Fatal("Must pass shard")
	}
}

func main() {
	parseFlags()

	// Parse sharding configuration
	c, err := config.ParseFile(*configFile)
	if err != nil {
		log.Fatalf("Error parsing config %q: %v", *configFile, err)
	}

	shards, err := config.ParseShards(c.Shards, *shard)
	if err != nil {
		log.Fatalf("Error parsing shards config: %v", err)
	}

	log.Printf("Shard count is %d, current shard: %d", shards.Count, shards.CurIdx)

	// Initialize storage
	store, err := storage.NewStorage(*dbLocation)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize server with storage, cache, and sharding
	srv := web.NewServer(store, shards, *cacheSize)

	// Register API routes
	http.HandleFunc("/api/v1/topics", srv.CreateTopicHandler)
	http.HandleFunc("/api/v1/topics/", srv.GetTopicHandler)
	http.HandleFunc("/api/v1/topics/", srv.StoreContextHandler)
	http.HandleFunc("/api/v1/topics/", srv.GetContextHandler)
	http.HandleFunc("/api/v1/topics/", srv.DeleteContextHandler)
	http.HandleFunc("/api/v1/topics", srv.ListTopicsHandler)
	http.HandleFunc("/api/v1/topics/", srv.ListContextsHandler)

	log.Printf("Starting server on %s", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
