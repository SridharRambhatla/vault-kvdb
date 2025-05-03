package server

import (
	"encoding/json"
	"net/http"
	"time"

	"go-kvdb/internal/cache"
	"go-kvdb/internal/storage"
	"go-kvdb/pkg/types"
)

// Server represents the HTTP server
type Server struct {
	storage *storage.Storage
	cache   *cache.LRUCache
}

// NewServer creates a new server instance
func NewServer(cacheSize int64) *Server {
	return &Server{
		storage: storage.NewStorage(),
		cache:   cache.NewLRUCache(cacheSize),
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	http.HandleFunc("/api/v1/topics", s.handleTopics)
	http.HandleFunc("/api/v1/topics/get", s.handleGetTopic)
	http.HandleFunc("/api/v1/topics/contexts/store", s.handleStoreContext)
	http.HandleFunc("/api/v1/topics/contexts/get", s.handleGetContext)
	http.HandleFunc("/api/v1/topics/contexts/delete", s.handleDeleteContext)
	http.HandleFunc("/api/v1/topics/list", s.handleListTopics)
	http.HandleFunc("/api/v1/topics/contexts/list", s.handleListContexts)

	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var topic types.Topic
	if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.storage.CreateTopic(&topic); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    topic,
	})
}

func (s *Server) handleGetTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic, err := s.storage.GetTopic(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    topic,
	})
}

func (s *Server) handleStoreContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ctx types.Context
	if err := json.NewDecoder(r.Body).Decode(&ctx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set timestamps
	now := time.Now()
	ctx.CreatedAt = now
	ctx.UpdatedAt = now

	// Store in both storage and cache
	if err := s.storage.StoreContext(&ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.cache.Put(ctx.ID, &ctx)

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    ctx,
	})
}

func (s *Server) handleGetContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Topic     string `json:"topic"`
		ContextID string `json:"context_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Try cache first
	ctx, err := s.cache.Get(req.ContextID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If not in cache, try storage
	if ctx == nil {
		ctx, err = s.storage.GetContext(req.ContextID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If found in storage, add to cache
		if ctx != nil {
			s.cache.Put(ctx.ID, ctx)
		}
	}

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    ctx,
	})
}

func (s *Server) handleDeleteContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Topic     string `json:"topic"`
		ContextID string `json:"context_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Delete from both storage and cache
	if err := s.storage.DeleteContext(req.ContextID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.cache.Remove(req.ContextID)

	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
	})
}

func (s *Server) handleListTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topics := s.storage.ListTopics()
	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    topics,
	})
}

func (s *Server) handleListContexts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	contexts := s.storage.ListContextsByTopic(req.Topic)
	json.NewEncoder(w).Encode(types.APIResponse{
		Success: true,
		Data:    contexts,
	})
}
