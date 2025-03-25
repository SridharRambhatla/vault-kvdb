package web

import (
	"fmt"
	"go-kvdb/db"
	"net/http"
)

type Server struct {
	db *db.Database
}

func NewServer(db *db.Database) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	bucketName := r.Form.Get("bucketName")

	value, err := s.db.GetKey(bucketName, key)
	fmt.Fprintf(w, "Value = %q, error = %v", value, err)
}

func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	value := r.Form.Get("value")
	bucketName := string(r.Form.Get("bucketName"))

	err := s.db.SetKey(key, bucketName, []byte(value))
	fmt.Fprintln(w, "Error = %v", err)
}

// Function to create a new bucket
func (s *Server) CreateBucket(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	bucketName := string(r.Form.Get("bucketName"))

	err := s.db.CreateBucketIfNotExists(bucketName)
	fmt.Fprintln(w, "Error = %v", err)
}
