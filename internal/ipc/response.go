package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/go-git/go-git/v5/plumbing"
)

func WriteErrorResponse(conn net.Conn) {
	message := []byte("error")
	conn.Write([]byte(fmt.Sprintf("%04x", len(message)+4)))
	conn.Write(message)
	conn.Write([]byte("0000")) // flush
}

func GitHashToObjectId(hash plumbing.Hash) (*ObjectId, error) {
	binHash, err := hex.DecodeString(hash.String())
	if err != nil {
		return nil, err
	}

	oid := &ObjectId{
		HashAlgo: 1,
	}
	copy(oid.Hash[:len(binHash)], binHash)
	return oid, nil
}

// TODO: struct packing
type GetOidResponse struct {
	Key
	Oid          ObjectId
	DeltaBaseOid ObjectId
	DiskSize     int64
	Size         uint32
	Whence       uint16
	Type         int8
	_            byte
}

func (resp *GetOidResponse) WriteResponse(conn net.Conn) error {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, resp)
	if err != nil {
		return err
	}

	fmt.Printf("Response size: %d\n", buf.Len())
	conn.Write([]byte(fmt.Sprintf("%04x", buf.Len()+4)))
	conn.Write(buf.Bytes())
	conn.Write([]byte("0000")) // flush
	return nil
}
