package db

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// The "Git" DB returns object database information using go-git
type gitDb struct {
	repo *git.Repository
}

func NewGitDb(path string) (Database, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit:          false,
		EnableDotGitCommonDir: false,
	})
	if err != nil {
		return nil, err
	}

	return &gitDb{
		repo: repo,
	}, nil
}

func (db *gitDb) ReadObject(oid plumbing.Hash, includeContent bool) (plumbing.EncodedObject, error) {
	backend, ok := db.repo.Storer.(storer.DeltaObjectStorer)
	if !ok {
		return nil, fmt.Errorf("cannot resolve delta object storage")
	}
	return backend.DeltaObject(plumbing.AnyObject, oid)
}
