package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"unsafe"

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
	_            [4]byte // WHY???
	Whence       uint32
	Type         int32
}

func (resp *GetOidResponse) WriteResponse(conn net.Conn, contentReader io.ReadCloser) error {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, resp)
	if err != nil {
		return err
	}

	responseSize := buf.Len() + 4
	if contentReader != nil {
		responseSize += int(resp.Size)
	}
	fmt.Printf("Response size: %d\n", responseSize)
	fmt.Printf("Struct size: %d\n", unsafe.Sizeof(GetOidResponse{}))

	var content []byte
	if contentReader != nil {
		content = make([]byte, resp.Size)
		n, err := contentReader.Read(content)
		if err != nil {
			return err
		} else if n != int(resp.Size) {
			return fmt.Errorf("mismatched content size (expected %d, got %d)", resp.Size, n)
		}
	}

	_, err = conn.Write([]byte(fmt.Sprintf("%04x", responseSize)))
	if err != nil {
		return err
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	if contentReader != nil {
		_, err = conn.Write(content)
		if err != nil {
			return err
		}
	}

	_, err = conn.Write([]byte("0000")) // flush
	if err != nil {
		return err
	}

	return nil
}
