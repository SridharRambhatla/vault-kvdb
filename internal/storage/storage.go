package storage

import (
	"sync"
	"time"

	"go-kvdb/pkg/types"
)

// Storage implements an in-memory storage for contexts
type Storage struct {
	mu       sync.RWMutex
	contexts map[string]*types.Context
	topics   map[string]*types.Topic
}

// NewStorage creates a new in-memory storage
func NewStorage() *Storage {
	return &Storage{
		contexts: make(map[string]*types.Context),
		topics:   make(map[string]*types.Topic),
	}
}

// StoreContext stores a context in memory
func (s *Storage) StoreContext(ctx *types.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure topic exists
	if _, ok := s.topics[ctx.Topic]; !ok {
		s.topics[ctx.Topic] = &types.Topic{
			Name:        ctx.Topic,
			Description: "Auto-created topic",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	// Store context
	s.contexts[ctx.ID] = ctx
	return nil
}

// GetContext retrieves a context by ID
func (s *Storage) GetContext(id string) (*types.Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ctx, ok := s.contexts[id]; ok {
		return ctx, nil
	}
	return nil, nil
}

// DeleteContext removes a context
func (s *Storage) DeleteContext(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.contexts, id)
	return nil
}

// ListContexts returns all contexts
func (s *Storage) ListContexts() []*types.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contexts := make([]*types.Context, 0, len(s.contexts))
	for _, ctx := range s.contexts {
		contexts = append(contexts, ctx)
	}
	return contexts
}

// ListContextsByTopic returns all contexts for a specific topic
func (s *Storage) ListContextsByTopic(topic string) []*types.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contexts := make([]*types.Context, 0)
	for _, ctx := range s.contexts {
		if ctx.Topic == topic {
			contexts = append(contexts, ctx)
		}
	}
	return contexts
}

// CreateTopic creates a new topic
func (s *Storage) CreateTopic(topic *types.Topic) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.topics[topic.Name]; ok {
		return nil // Topic already exists
	}

	s.topics[topic.Name] = topic
	return nil
}

// GetTopic retrieves a topic by name
func (s *Storage) GetTopic(name string) (*types.Topic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if topic, ok := s.topics[name]; ok {
		return topic, nil
	}
	return nil, nil
}

// ListTopics returns all topics
func (s *Storage) ListTopics() []*types.Topic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	topics := make([]*types.Topic, 0, len(s.topics))
	for _, topic := range s.topics {
		topics = append(topics, topic)
	}
	return topics
}
