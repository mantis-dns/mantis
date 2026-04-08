package storage

import (
	"fmt"

	"github.com/cockroachdb/pebble/v2"
)

// DB wraps a Pebble database instance.
type DB struct {
	pdb *pebble.DB
}

// Open creates or opens a Pebble database at the given path.
func Open(path string) (*DB, error) {
	opts := &pebble.Options{
		// Use default options; tune later based on profiling.
	}
	pdb, err := pebble.Open(path, opts)
	if err != nil {
		return nil, fmt.Errorf("open pebble db: %w", err)
	}
	return &DB{pdb: pdb}, nil
}

// Close closes the Pebble database.
func (db *DB) Close() error {
	return db.pdb.Close()
}

// Pebble returns the underlying pebble.DB for direct access.
func (db *DB) Pebble() *pebble.DB {
	return db.pdb
}
