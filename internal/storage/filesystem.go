package storage

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

// "Filesystem" storage is the default Git on-disk storage
func NewFilesystemStorage(path string) GitStorage {
	return &filesystemStorage{
		Storage: filesystem.NewStorage(osfs.New(path), cache.NewObjectLRUDefault()),
	}
}

type filesystemStorage struct {
	*filesystem.Storage
}

func (*filesystemStorage) Close() error {
	return nil // no op
}
