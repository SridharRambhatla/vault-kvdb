package web

import (
	"fmt"
	"go-kvdb/db"
	"hash/fnv"
	"net/http"
)

type Server struct {
	db         *db.Database
	shardIdx   int
	shardCount int
}

func NewServer(db *db.Database, shardCount, shardIdx int) *Server {
	return &Server{
		db:         db,
		shardIdx:   shardIdx,
		shardCount: shardCount,
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

	h := fnv.New64()
	h.Write([]byte(key))
	shardIdx := int(h.Sum64() % uint64(s.shardCount))

	err := s.db.SetKey(key, bucketName, []byte(value))
	fmt.Fprintf(w, "Error = %v, hash = %d, shardIdx = %d", err, h.Sum64(), shardIdx)
}

// Function to create a new bucket
func (s *Server) CreateBucket(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	bucketName := string(r.Form.Get("bucketName"))

	err := s.db.CreateBucketIfNotExists(bucketName)
	fmt.Fprintf(w, "Error = %v", err)
}
