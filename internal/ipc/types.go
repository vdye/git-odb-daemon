package ipc

import (
	"bytes"
	"encoding/hex"

	"github.com/go-git/go-git/v5/plumbing"
)

type RequestSizeString [4]byte

type Key [16]byte

func (k *Key) ToString() string {
	return string(bytes.Trim(k[:], "\x00"))
}

type ObjectId struct {
	Hash     [32]byte
	HashAlgo int32
}

func (oid *ObjectId) Hex() string {
	if oid.HashAlgo == 1 {
		return hex.EncodeToString(oid.Hash[:20])
	} else {
		return hex.EncodeToString(oid.Hash[:])
	}
}

func (oid *ObjectId) GitHash() plumbing.Hash {
	return plumbing.NewHash(oid.Hex())
}
