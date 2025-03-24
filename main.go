package main

import (
	"flag"
	"fmt"
	"go-kvdb/db"
	"log"
	"net/http"
)

var (
	dbLocation = flag.String("db-location", "", "The path to the bolt db database")
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

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		key := r.Form.Get("key")
		bucket := r.Form.Get("bucket")

		value, err := db.getKey(bucket, key)
		fmt.Fprintf(w, "Value = %q, error = %v", value, err)
	})
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		key := r.Form.Get("key")
		value := r.Form.Get("value")
		bucket := r.Form.Get("bucket")

		err := db.setKey(key, bucket, []byte(value))
		fmt.Fprint(w, "Error = %v", err)
	})
}
