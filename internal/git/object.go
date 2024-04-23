package git

import "github.com/vdye/git-odb-daemon/internal/types"

type Object interface {
	Oid() (types.ObjectId, error)
	RawGitBuffer() ([]byte, error)
}

type ObjectPointer interface {
	Oid() types.ObjectId
	GetObject() (Object, error)
}
