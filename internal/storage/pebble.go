package storage

import (
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage"
)

func NewPebbleStorage(path string) (GitStorage, error) {
	conn, err := pebble.Open(filepath.Join(path, "objects", "pebble"), &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &PebbleStorage{
		conn: conn,
	}, nil
}

type PebbleStorage struct {
	conn *pebble.DB // Object storage

	// Everything else (nil for now, since we shouldn't be using any of them)
	storer.ReferenceStorer
	storer.ShallowStorer
	storer.IndexStorer
	config.ConfigStorer
	storage.ModuleStorer
}

func (s *PebbleStorage) Close() error {
	return s.conn.Close()
}

func (s *PebbleStorage) NewEncodedObject() plumbing.EncodedObject {
	return nil // Not implemented
}

func (s *PebbleStorage) SetEncodedObject(obj plumbing.EncodedObject) (plumbing.Hash, error) {
	return plumbing.ZeroHash, fmt.Errorf("not implemented")
}

func (s *PebbleStorage) EncodedObject(objType plumbing.ObjectType, oid plumbing.Hash) (plumbing.EncodedObject, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *PebbleStorage) IterEncodedObjects(objType plumbing.ObjectType) (storer.EncodedObjectIter, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *PebbleStorage) HasEncodedObject(oid plumbing.Hash) error {
	return fmt.Errorf("not implemented")
}

func (s *PebbleStorage) EncodedObjectSize(oid plumbing.Hash) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (s *PebbleStorage) AddAlternate(remote string) error {
	// No alternates support
	return fmt.Errorf("alternates are not supported")
}
