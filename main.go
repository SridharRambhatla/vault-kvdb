package main

import (
	"flag"
	"go-kvdb/config"
	"go-kvdb/db"
	"go-kvdb/web"
	"log"
	"net/http"
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
		log.Fatalf("Must provide db-location")
	}

	if *shard == "" {
		log.Fatalf("Must provide shard")
	}
}

func main() {
	parseFlags()

	c, err := config.ParseFile(*configFile)
	if err != nil {
		log.Fatalf("Error parsing config %q: %v", *configFile, err)
	}

	shards, err := config.ParseShards(c.Shards, *shard)
	if err != nil {
		log.Fatalf("Error parsing shards config: %v", err)
	}

	log.Printf("Shard count is %d, current shard: %d", shards.Count, shards.CurIdx)

	db, close, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatalf("Error creating %q: %v", *dbLocation, err)
	}
	defer close()

	srv := web.NewServer(db, shards)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)
	http.HandleFunc("/purge", srv.DeleteExtraKeysHandler)
	http.HandleFunc("/createBucket", srv.CreateBucket)
	http.HandleFunc("/deleteBucket", srv.DeleteBucketHandler)
	http.HandleFunc("/listKeys", srv.ListKeysHandler)

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
