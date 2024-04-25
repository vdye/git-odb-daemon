package storage

import "github.com/go-git/go-git/v5/plumbing/storer"

type GitStorage interface {
	storer.Storer
	Close() error
}
