package main

import (
	"flag"
	"go-kvdb/config"
	"go-kvdb/db"
	"go-kvdb/web"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
)

var (
	dbLocation = flag.String("db-location", "", "The path to the bolt db database")
	httpAddr   = flag.String("http-addr", "127.0.0.1:8080", "HTTP host and port")
	configFile = flag.String("config-file", "sharding.toml", "Config file for static sharding")
	shard      = flag.String("shard", "", "The name of the shard for data")
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

	var c config.Config
	if _, err := toml.DecodeFile(*configFile, &c); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var shardCount int
	var shardIdx int
	shardCount = len(c.Shards)
	for _, s := range c.Shards {
		if s.Name == *shard {
			shardIdx = s.Idx
		}
	}
	if shardIdx < 0 {
		log.Fatalf("Shard %q was not found", *shard)
	}

	log.Printf("Total shard count is %d, current shard key is - %d", shardCount, shardIdx)

	db, close, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("Something went wrong %q %v", *dbLocation, err)
	}

	defer close()

	srv := web.NewServer(db, shardCount, shardIdx)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)
	http.HandleFunc("/createBucket", srv.CreateBucket)

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
