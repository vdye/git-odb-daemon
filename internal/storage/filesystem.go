package storage

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

// "Filesystem" storage is the default Git on-disk storage
func NewFilesystemStorage(path string) storage.Storer {
	return filesystem.NewStorage(osfs.New(path), cache.NewObjectLRUDefault())
}
