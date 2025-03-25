package main

import (
	"flag"
	"go-kvdb/db"
	"go-kvdb/web"
	"log"
	"net/http"
)

var (
	dbLocation = flag.String("db-location", "", "The path to the bolt db database")
	httpAddr   = flag.String("http-addr", "127.0.0.1:8080", "HTTP host and port")
)

func parseFlags() {
	flag.Parse()

	if *dbLocation == "" {
		log.Fatal("Must pass db-location")
	}
}

func main() {
	parseFlags()

	db, close, err := db.NewDatabase(*dbLocation)
	if err != nil {
		log.Fatal("Something went wrong %q %v", *dbLocation, err)
	}

	defer close()

	srv := web.NewServer(db)

	http.HandleFunc("/get", srv.GetHandler)
	http.HandleFunc("/set", srv.SetHandler)

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
