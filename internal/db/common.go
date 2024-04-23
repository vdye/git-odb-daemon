package db

import (
	"github.com/go-git/go-git/v5/plumbing"
)

type Database interface {
	ReadObject(oid plumbing.Hash, includeContent bool) (plumbing.EncodedObject, error)
}
