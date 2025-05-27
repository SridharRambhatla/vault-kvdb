package db_test

import (
	"bytes"
	"go-kvdb/db"
	"io/ioutil"
	"os"
	"slices"
	"testing"
)

func TestGetSet(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "database")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Could not create a new database: %v", err)
	}
	defer closeFunc()

	if err := db.SetKey("party", "default", []byte("Great")); err != nil {
		t.Fatalf("Could not write key: %v", err)
	}

	value, err := db.GetKey("party", "default")
	if err != nil {
		t.Fatalf(`Could not get the key "party": %v`, err)
	}

	if !bytes.Equal(value, []byte("Great")) {
		t.Errorf(`Unexpected value for key "party": got %q, want %q`, value, "Great")
	}
}

func setKey(t *testing.T, d *db.Database, key, value string, bucketName string) {
	t.Helper()

	if err := d.SetKey(key, bucketName, []byte(value)); err != nil {
		t.Fatalf("SetKey(%q, %q) failed: %v", key, value, err)
	}
}

func getKey(t *testing.T, d *db.Database, key string, bucketName string) string {
	t.Helper()

	value, err := d.GetKey(key, bucketName)
	if err != nil {
		t.Fatalf("GetKey(%q) failed: %v", key, err)
	}

	return string(value)
}

func TestDeleteExtraKeys(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "database")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Could not create a new database: %v", err)
	}
	defer closeFunc()

	setKey(t, db, "party", "Great", "default")
	setKey(t, db, "us", "CapitalistPigs", "default")

	if err := db.DeleteExtraKeys(func(name string) bool { return name == "us" }, "default"); err != nil {
		t.Fatalf("Could not delete extra keys: %v", err)
	}

	if value := getKey(t, db, "party", "default"); value != "Great" {
		t.Errorf(`Unexpected value for key "party": got %q, want %q`, value, "Great")
	}

	if value := getKey(t, db, "us", "default"); value != "" {
		t.Errorf(`Unexpected value for key "us": got %q, want %q`, value, "")
	}
}

func TestListKeys(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "database")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Could not create a new database: %v", err)
	}
	defer closeFunc()

	setKey(t, db, "party", "Great", "default")
	setKey(t, db, "us", "CapitalistPigs", "default")

	keys, err := db.ListKeys("default")
	t.Logf("keys: %v", keys)
	if err != nil {
		t.Fatalf("Could not list keys: %v", err)
	}

	if len(keys) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(keys))
	}

	if !slices.Contains(keys, "party") {
		t.Errorf("Expected key 'party' to be in the list")
	}

	if !slices.Contains(keys, "us") {
		t.Errorf("Expected key 'us' to be in the list")
	}
}

func TestCreateBucket(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "database")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Could not create a new database: %v", err)
	}
	defer closeFunc()

	// Test creating a new bucket
	if err := db.CreateBucketIfNotExists("test_bucket"); err != nil {
		t.Fatalf("Could not create bucket: %v", err)
	}

	// Test creating the same bucket again (should not error)
	if err := db.CreateBucketIfNotExists("test_bucket"); err != nil {
		t.Fatalf("Could not create existing bucket: %v", err)
	}

	// Test that we can use the bucket
	if err := db.SetKey("test_key", "test_bucket", []byte("test_value")); err != nil {
		t.Fatalf("Could not set key in new bucket: %v", err)
	}

	value, err := db.GetKey("test_key", "test_bucket")
	if err != nil {
		t.Fatalf("Could not get key from new bucket: %v", err)
	}

	if string(value) != "test_value" {
		t.Errorf("Expected value 'test_value', got %q", value)
	}

	// Test listing keys in the new bucket
	keys, err := db.ListKeys("test_bucket")
	if err != nil {
		t.Fatalf("Could not list keys: %v", err)
	}

	if len(keys) != 1 {
		t.Fatalf("Expected 1 key, got %d", len(keys))
	}

	if !slices.Contains(keys, "test_key") {
		t.Errorf("Expected key 'test_key' to be in the list")
	}
}

func TestDeleteBucket(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "database")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)

	db, closeFunc, err := db.NewDatabase(name)
	if err != nil {
		t.Fatalf("Could not create a new database: %v", err)
	}
	defer closeFunc()

	// Create a test bucket
	if err := db.CreateBucketIfNotExists("test_bucket"); err != nil {
		t.Fatalf("Could not create bucket: %v", err)
	}

	// Add some data to the bucket
	if err := db.SetKey("test_key", "test_bucket", []byte("test_value")); err != nil {
		t.Fatalf("Could not set key in bucket: %v", err)
	}

	// Delete the bucket
	if err := db.DeleteBucket("test_bucket"); err != nil {
		t.Fatalf("Could not delete bucket: %v", err)
	}

	// Verify bucket is deleted by trying to get a key from it
	_, err = db.GetKey("test_key", "test_bucket")
	if err == nil {
		t.Error("Expected error when getting key from deleted bucket")
	}

	// Try to delete non-existent bucket
	err = db.DeleteBucket("non_existent_bucket")
	if err == nil {
		t.Error("Expected error when deleting non-existent bucket")
	}
}
