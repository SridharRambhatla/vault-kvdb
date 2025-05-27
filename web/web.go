package web

import (
	"fmt"
	"go-kvdb/config"
	"go-kvdb/db"
	"io"
	"net/http"
)

// Server contains HTTP method handlers to be used for the database.
type Server struct {
	db     *db.Database
	shards *config.Shards
}

// NewServer creates a new instance with HTTP handlers to be used to get and set values.
func NewServer(db *db.Database, s *config.Shards) *Server {
	return &Server{
		db:     db,
		shards: s,
	}
}

func (s *Server) redirect(shard int, w http.ResponseWriter, r *http.Request) {
	url := "http://" + s.shards.Addrs[shard] + r.RequestURI
	fmt.Fprintf(w, "redirecting from shard %d to shard %d (%q)\n", s.shards.CurIdx, shard, url)

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error redirecting the request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Error from target shard: %s", resp.Status), resp.StatusCode)
		return
	}

	io.Copy(w, resp.Body)
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	key := r.Form.Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	bucketName := r.Form.Get("bucketName")
	if bucketName == "" {
		bucketName = "default"
	}

	shard := s.shards.Index(key)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	value, err := s.db.GetKey(key, bucketName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting key: %v", err), http.StatusInternalServerError)
		return
	}

	if value == nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Shard = %d, current shard = %d, addr = %q, Value = %q",
		shard, s.shards.CurIdx, s.shards.Addrs[shard], value)
}

func (s *Server) SetHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	key := r.Form.Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	value := r.Form.Get("value")
	if value == "" {
		http.Error(w, "value parameter is required", http.StatusBadRequest)
		return
	}

	bucketName := r.Form.Get("bucketName")
	if bucketName == "" {
		bucketName = "default"
	}

	shard := s.shards.Index(key)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	err := s.db.SetKey(key, bucketName, []byte(value))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error setting key: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully set key in shard %d", shard)
}

// Function to create a new bucket
func (s *Server) CreateBucket(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	bucketName := r.Form.Get("bucketName")
	if bucketName == "" {
		http.Error(w, "bucketName parameter is required", http.StatusBadRequest)
		return
	}

	err := s.db.CreateBucketIfNotExists(bucketName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating bucket: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully created bucket %s", bucketName)
}

// DeleteExtraKeysHandler deletes keys that don't belong to the current shard.
func (s *Server) DeleteExtraKeysHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	bucketName := r.Form.Get("bucketName")
	if bucketName == "" {
		bucketName = "default"
	}

	err := s.db.DeleteExtraKeys(func(key string) bool {
		return s.shards.Index(key) != s.shards.CurIdx
	}, bucketName)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting extra keys: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully deleted extra keys from bucket %s", bucketName)
}

// DeleteBucketHandler handles the deletion of a bucket.
func (s *Server) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	bucketName := r.Form.Get("bucketName")
	if bucketName == "" {
		http.Error(w, "bucketName parameter is required", http.StatusBadRequest)
		return
	}

	// Don't allow deletion of the default bucket
	if bucketName == "default" {
		http.Error(w, "Cannot delete the default bucket", http.StatusBadRequest)
		return
	}

	err := s.db.DeleteBucket(bucketName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting bucket: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully deleted bucket %s", bucketName)
}

// ListKeysHandler returns all keys in the specified bucket.
func (s *Server) ListKeysHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	bucketName := r.Form.Get("bucketName")
	if bucketName == "" {
		bucketName = "default"
	}

	keys, err := s.db.ListKeys(bucketName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing keys: %v", err), http.StatusInternalServerError)
		return
	}

	if len(keys) == 0 {
		fmt.Fprintf(w, "No keys found in bucket %s", bucketName)
		return
	}

	fmt.Fprintf(w, "Keys in bucket %s:\n", bucketName)
	for _, key := range keys {
		fmt.Fprintf(w, "- %s\n", key)
	}
}
