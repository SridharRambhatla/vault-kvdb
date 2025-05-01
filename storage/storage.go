package storage

import (
	"encoding/json"
	"errors"
	"time"

	"go-kvdb/types"
	"go-kvdb/utils"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

var (
	ErrTopicNotFound   = errors.New("topic not found")
	ErrContextNotFound = errors.New("context not found")
)

// Storage implements the storage layer using BoltDB
type Storage struct {
	db *bbolt.DB
}

// NewStorage creates a new storage instance
func NewStorage(dbPath string) (*Storage, error) {
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create necessary buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("topics"))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// CreateTopic creates a new topic
func (s *Storage) CreateTopic(name string) (*types.Topic, error) {
	topic := &types.Topic{
		Name:         name,
		CreatedAt:    time.Now(),
		LastUpdated:  time.Now(),
		ContextCount: 0,
	}

	err := s.db.Update(func(tx *bbolt.Tx) error {
		topics := tx.Bucket([]byte("topics"))
		if topics == nil {
			return errors.New("topics bucket not found")
		}

		// Create topic bucket
		_, err := topics.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}

		// Store topic metadata
		topicData, err := json.Marshal(topic)
		if err != nil {
			return err
		}

		return topics.Put([]byte(name+"_meta"), topicData)
	})

	if err != nil {
		return nil, err
	}

	return topic, nil
}

// GetTopic retrieves a topic
func (s *Storage) GetTopic(name string) (*types.Topic, error) {
	var topic types.Topic

	err := s.db.View(func(tx *bbolt.Tx) error {
		topics := tx.Bucket([]byte("topics"))
		if topics == nil {
			return errors.New("topics bucket not found")
		}

		topicData := topics.Get([]byte(name + "_meta"))
		if topicData == nil {
			return ErrTopicNotFound
		}

		return json.Unmarshal(topicData, &topic)
	})

	if err != nil {
		return nil, err
	}

	return &topic, nil
}

// StoreContext stores a new context in a topic
func (s *Storage) StoreContext(topicName string, context *types.Context) error {
	if context.Metadata.ID == "" {
		context.Metadata.ID = uuid.New().String()
	}

	context.Metadata.CreatedAt = time.Now()
	context.Metadata.LastAccessed = time.Now()
	context.Metadata.Topic = topicName

	// Compress the content
	compressed, err := utils.CompressAndMarshal(context.Content)
	if err != nil {
		return err
	}

	context.Content.Data = compressed
	context.Content.Compressed = true

	return s.db.Update(func(tx *bbolt.Tx) error {
		topics := tx.Bucket([]byte("topics"))
		if topics == nil {
			return errors.New("topics bucket not found")
		}

		topic := topics.Bucket([]byte(topicName))
		if topic == nil {
			return ErrTopicNotFound
		}

		// Store context
		contextData, err := json.Marshal(context)
		if err != nil {
			return err
		}

		if err := topic.Put([]byte(context.Metadata.ID), contextData); err != nil {
			return err
		}

		// Update topic metadata
		var topicMeta types.Topic
		topicData := topics.Get([]byte(topicName + "_meta"))
		if err := json.Unmarshal(topicData, &topicMeta); err != nil {
			return err
		}

		topicMeta.ContextCount++
		topicMeta.LastUpdated = time.Now()

		topicData, err = json.Marshal(topicMeta)
		if err != nil {
			return err
		}

		return topics.Put([]byte(topicName+"_meta"), topicData)
	})
}

// GetContext retrieves a context from a topic
func (s *Storage) GetContext(topicName, contextID string) (*types.Context, error) {
	var context types.Context

	err := s.db.View(func(tx *bbolt.Tx) error {
		topics := tx.Bucket([]byte("topics"))
		if topics == nil {
			return errors.New("topics bucket not found")
		}

		topic := topics.Bucket([]byte(topicName))
		if topic == nil {
			return ErrTopicNotFound
		}

		contextData := topic.Get([]byte(contextID))
		if contextData == nil {
			return ErrContextNotFound
		}

		if err := json.Unmarshal(contextData, &context); err != nil {
			return err
		}

		// Decompress the content if it's compressed
		if context.Content.Compressed {
			var decompressed interface{}
			if err := utils.UnmarshalAndDecompress(context.Content.Data.([]byte), &decompressed); err != nil {
				return err
			}
			context.Content.Data = decompressed
			context.Content.Compressed = false
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Update last accessed time
	context.Metadata.LastAccessed = time.Now()
	if err := s.StoreContext(topicName, &context); err != nil {
		return nil, err
	}

	return &context, nil
}

// DeleteContext deletes a context from a topic
func (s *Storage) DeleteContext(topicName, contextID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		topics := tx.Bucket([]byte("topics"))
		if topics == nil {
			return errors.New("topics bucket not found")
		}

		topic := topics.Bucket([]byte(topicName))
		if topic == nil {
			return ErrTopicNotFound
		}

		if err := topic.Delete([]byte(contextID)); err != nil {
			return err
		}

		// Update topic metadata
		var topicMeta types.Topic
		topicData := topics.Get([]byte(topicName + "_meta"))
		if err := json.Unmarshal(topicData, &topicMeta); err != nil {
			return err
		}

		topicMeta.ContextCount--
		topicMeta.LastUpdated = time.Now()

		topicData, err := json.Marshal(topicMeta)
		if err != nil {
			return err
		}

		return topics.Put([]byte(topicName+"_meta"), topicData)
	})
}

// ListTopics lists all topics
func (s *Storage) ListTopics() ([]*types.Topic, error) {
	var topics []*types.Topic

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("topics"))
		if b == nil {
			return errors.New("topics bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			if string(k[len(k)-5:]) == "_meta" {
				var topic types.Topic
				if err := json.Unmarshal(v, &topic); err != nil {
					return err
				}
				topics = append(topics, &topic)
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return topics, nil
}

// ListContexts lists all contexts in a topic
func (s *Storage) ListContexts(topicName string) ([]*types.Context, error) {
	var contexts []*types.Context

	err := s.db.View(func(tx *bbolt.Tx) error {
		topics := tx.Bucket([]byte("topics"))
		if topics == nil {
			return errors.New("topics bucket not found")
		}

		topic := topics.Bucket([]byte(topicName))
		if topic == nil {
			return ErrTopicNotFound
		}

		return topic.ForEach(func(k, v []byte) error {
			if string(k[len(k)-5:]) != "_meta" {
				var context types.Context
				if err := json.Unmarshal(v, &context); err != nil {
					return err
				}

				// Decompress the content if it's compressed
				if context.Content.Compressed {
					var decompressed interface{}
					if err := utils.UnmarshalAndDecompress(context.Content.Data.([]byte), &decompressed); err != nil {
						return err
					}
					context.Content.Data = decompressed
					context.Content.Compressed = false
				}

				contexts = append(contexts, &context)
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return contexts, nil
}
