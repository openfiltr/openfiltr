package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestOpenBoltBootstrapsFreshDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openfiltr.db")

	store, err := OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database file missing after open: %v", err)
	}

	if err := store.View(func(tx *bolt.Tx) error {
		meta := tx.Bucket([]byte(bboltMetadataBucket))
		if meta == nil {
			return fmt.Errorf("metadata bucket missing")
		}
		got := string(meta.Get([]byte(bboltVersionKey)))
		if got != strconv.Itoa(bboltStoreVersion) {
			return fmt.Errorf("store version = %q, want %d", got, bboltStoreVersion)
		}
		for _, name := range bboltRequiredBuckets {
			if tx.Bucket([]byte(name)) == nil {
				return fmt.Errorf("bucket %q missing", name)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("view failed: %v", err)
	}
}

func TestOpenBoltPersistsRecordsAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openfiltr.db")

	store, err := OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}

	if err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bboltMetadataBucket))
		if b == nil {
			return fmt.Errorf("metadata bucket missing")
		}
		return b.Put([]byte("sample"), []byte("value"))
	}); err != nil {
		t.Fatalf("writing record = %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := OpenBolt(path)
	if err != nil {
		t.Fatalf("reopening database failed: %v", err)
	}
	defer reopened.Close()

	if err := reopened.View(func(tx *bolt.Tx) error {
		meta := tx.Bucket([]byte(bboltMetadataBucket))
		if meta == nil {
			return fmt.Errorf("metadata bucket missing after reopen")
		}
		got := string(meta.Get([]byte(bboltVersionKey)))
		if got != strconv.Itoa(bboltStoreVersion) {
			return fmt.Errorf("store version = %q, want %d", got, bboltStoreVersion)
		}
		if string(meta.Get([]byte("sample"))) != "value" {
			return fmt.Errorf("stored record did not survive reopen")
		}
		return nil
	}); err != nil {
		t.Fatalf("view after reopen failed: %v", err)
	}
}
