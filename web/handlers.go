package web

import (
	"encoding/json"
	"fmt"
	"go-kvdb/cache"
	"go-kvdb/config"
	"go-kvdb/storage"
	"go-kvdb/types"
	"io"
	"net/http"
	"strings"
)

// Server represents the HTTP server
type Server struct {
	storage *storage.Storage
	shards  *config.Shards
	cache   *cache.LRUCache
}

// NewServer creates a new server instance
func NewServer(storage *storage.Storage, shards *config.Shards, cacheSize int64) *Server {
	return &Server{
		storage: storage,
		shards:  shards,
		cache:   cache.NewLRUCache(cacheSize),
	}
}

func (s *Server) redirect(shard int, w http.ResponseWriter, r *http.Request) {
	url := "http://" + s.shards.Addrs[shard] + r.RequestURI
	fmt.Fprintf(w, "redirecting from shard %d to shard %d (%q)\n", s.shards.CurIdx, shard, url)

	resp, err := http.Get(url)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error redirecting the request: %v", err)
		return
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)
}

// CreateTopicHandler handles topic creation
func (s *Server) CreateTopicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if we should handle this request
	shard := s.shards.Index(req.Name)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	topic, err := s.storage.CreateTopic(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    topic,
	})
}

// GetTopicHandler handles topic retrieval
func (s *Server) GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicName := strings.TrimPrefix(r.URL.Path, "/api/v1/topics/")

	// Check if we should handle this request
	shard := s.shards.Index(topicName)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	topic, err := s.storage.GetTopic(topicName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    topic,
	})
}

// StoreContextHandler handles context storage
func (s *Server) StoreContextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	topicName := pathParts[3]
	var context types.Context

	if err := json.NewDecoder(r.Body).Decode(&context); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if we should handle this request
	shard := s.shards.Index(context.Metadata.ID)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	if err := s.storage.StoreContext(topicName, &context); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update cache
	s.cache.Put(context.Metadata.ID, &context)

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    context,
	})
}

// GetContextHandler handles context retrieval
func (s *Server) GetContextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	topicName := pathParts[3]
	contextID := pathParts[4]

	// Check if we should handle this request
	shard := s.shards.Index(contextID)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	// Try cache first
	context, err := s.cache.Get(contextID)
	if err == nil {
		json.NewEncoder(w).Encode(types.APIResponse{
			Success: true,
			Data:    context,
		})
		return
	}

	// If not in cache, get from storage
	context, err = s.storage.GetContext(topicName, contextID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Update cache
	s.cache.Put(contextID, context)

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    context,
	})
}

// DeleteContextHandler handles context deletion
func (s *Server) DeleteContextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	topicName := pathParts[3]
	contextID := pathParts[4]

	// Check if we should handle this request
	shard := s.shards.Index(contextID)
	if shard != s.shards.CurIdx {
		s.redirect(shard, w, r)
		return
	}

	if err := s.storage.DeleteContext(topicName, contextID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove from cache
	s.cache.Remove(contextID)

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
	})
}

// ListTopicsHandler handles listing all topics
func (s *Server) ListTopicsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topics, err := s.storage.ListTopics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    topics,
	})
}

// ListContextsHandler handles listing all contexts in a topic
func (s *Server) ListContextsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	topicName := pathParts[3]
	contexts, err := s.storage.ListContexts(topicName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    contexts,
	})
}
