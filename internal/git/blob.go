package git

import (
	"crypto/sha1"

	"github.com/vdye/git-odb-daemon/internal/types"
)

type blob struct {
	oid    types.ObjectId
	buffer []byte
}

func NewBlob(buffer []byte) (Object, error) {
	hasher := sha1.New()
	_, err := hasher.Write(buffer)
	if err != nil {
		return nil, err
	}

	return &blob{
		// Hardcode SHA1 hash
		oid: types.ObjectId{
			Hash:     [32]byte(hasher.Sum(nil)),
			HashAlgo: 1,
		},
		buffer: buffer,
	}, nil
}

func (b *blob) Oid() (types.ObjectId, error) {
	return b.oid, nil
}

func (b *blob) RawGitBuffer() ([]byte, error) {
	return b.buffer, nil
}
