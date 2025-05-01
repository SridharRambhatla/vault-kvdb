package types

import (
	"time"
)

// ContextType represents the type of stored context
type ContextType string

const (
	ConversationType ContextType = "conversation"
	CodeType         ContextType = "code"
	DocType          ContextType = "doc"
)

// Metadata contains metadata for all stored items
type Metadata struct {
	CreatedAt    time.Time   `json:"created_at"`
	LastAccessed time.Time   `json:"last_accessed"`
	Size         int64       `json:"size"`
	Type         ContextType `json:"type"`
	Topic        string      `json:"topic"`
	ID           string      `json:"id"`
}

// Content represents the actual stored data
type Content struct {
	Data       interface{} `json:"data"`
	Compressed bool        `json:"compressed"`
}

// Context represents a complete context entry
type Context struct {
	Metadata Metadata `json:"metadata"`
	Content  Content  `json:"content"`
}

// Topic represents a collection of contexts
type Topic struct {
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	LastUpdated  time.Time `json:"last_updated"`
	ContextCount int       `json:"context_count"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
