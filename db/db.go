package db

import bolt "github.com/boltdb/bolt"

type Database struct {
	db *bolt.DB
}

func NewDatabase(path string) (db *Database, closeFunc func() error, err error) {
	boltdb, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, nil, err
	}

	closeFunc = boltdb.Close

	return &Database{db: boltdb}, closeFunc, nil
}

// Set key function
func (d *Database) setKey(key string, bucket, value []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(bucket))
		return b.Put([]byte(key), value)
	})
}

// Get key function
func (d *Database) getKey(bucket, key string) ([]byte, error) {
	var result []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		result = b.Get([]byte(key))
		return nil
	})
	if err == nil {
		return result, nil
	}
	return nil, err
}
