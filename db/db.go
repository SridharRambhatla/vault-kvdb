package db

import (
	"fmt"

	"github.com/boltdb/bolt"
)

// Database is an open bolt database.
type Database struct {
	db *bolt.DB
}

// NewDatabase returns an instance of a database that we can work with.
func NewDatabase(dbPath string) (db *Database, closeFunc func() error, err error) {
	boltDb, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, nil, err
	}

	db = &Database{db: boltDb}
	closeFunc = boltDb.Close

	// Optionally create default bucket
	if err := db.CreateBucketIfNotExists("default"); err != nil {
		closeFunc()
		return nil, nil, fmt.Errorf("creating default bucket: %w", err)
	}

	return db, closeFunc, nil
}

// createBucketIfNotExists creates a bucket in the database if it doesn't exist.
func (d *Database) CreateBucketIfNotExists(bucketName string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
}

// SetKey sets the key to the requested value in the specified bucket.
func (d *Database) SetKey(key string, bucketName string, value []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.Put([]byte(key), value)
	})
}

// GetKey gets the value of the requested key from the specified bucket.
func (d *Database) GetKey(bucketName string, key string) ([]byte, error) {
	var result []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		result = b.Get([]byte(key))
		return nil
	})

	if err == nil {
		return result, nil
	}
	return nil, err
}

// DeleteKey
func (d *Database) DelKey(bucketName string, key string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		return b.Delete(([]byte(key)))
	})
}

// ListKeys in a sorted manner
// DeleteBucket
