package ipc

import (
	"bytes"
	"encoding/hex"
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

type ObjectType uint8
