package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"unsafe"
)

type IpcRequest interface {
	Key() string
}

type EOF struct{}

func (*EOF) Key() string {
	return "EOF" // TODO: what if an IPC request is created with ID "EOF"
}

type FlushPacket struct{}

func (*FlushPacket) Key() string {
	return "flush" // TODO: what if an IPC request is created with ID "flush"
}

type GetOidRequest struct {
	ObjectId
	Flags       uint32
	WantContent uint8
}

func (*GetOidRequest) Key() string {
	return "oid"
}

type hashObjectRequestInternal struct {
	Type  int32
	Flags uint32
	Size  uint64
}

type HashObjectRequest struct {
	hashObjectRequestInternal
	Content []byte
}

func (*HashObjectRequest) Key() string {
	return "hash-object"
}

func ReadRequest(conn net.Conn) (IpcRequest, error) {
	// First, read the size of the request
	pktLine := make([]byte, 4)
	n, err := io.ReadFull(conn, pktLine)
	if err == io.EOF {
		return &EOF{}, nil
	} else if err != nil {
		return nil, err
	} else if n < 4 {
		return nil, fmt.Errorf("could not read size")
	}

	reqSize, err := strconv.ParseUint(string(pktLine), 16, 32)
	if err != nil {
		return nil, err
	}

	// Flush packet
	if reqSize == 0 {
		return &FlushPacket{}, nil
	}

	reqSize -= 4
	fmt.Printf("Request size: %d\n", reqSize)
	fmt.Printf("Struct size %d\n", unsafe.Sizeof(GetOidRequest{})+16)

	reqBuf := bytes.NewBuffer(make([]byte, reqSize))
	n, err = io.ReadFull(conn, reqBuf.Bytes())
	if err != nil {
		return nil, err
	} else if n < int(reqSize) {
		return nil, fmt.Errorf("request too small (expected %d, received %d)", reqSize, n)
	}

	// Look for flush packet
	// First, read the size of the request
	n, err = io.ReadFull(conn, pktLine)
	if err != nil {
		return nil, err
	} else if n < 4 {
		return nil, fmt.Errorf("could not read flush packet")
	} else if string(pktLine) != "0000" {
		return nil, fmt.Errorf("invalid packet line %s", string(pktLine))
	}

	var k Key
	err = binary.Read(reqBuf, binary.LittleEndian, &k) // TODO: use system endianness
	if err != nil {
		return nil, err
	}

	switch key := k.ToString(); key {
	case "oid":
		var oidReq GetOidRequest
		err = binary.Read(reqBuf, binary.LittleEndian, &oidReq)
		if err != nil {
			return nil, err
		}
		return &oidReq, nil
	case "hash-object":
		var internalReq hashObjectRequestInternal

		// Read the fixed-size component of the buffer first
		err = binary.Read(reqBuf, binary.LittleEndian, &internalReq)
		if err != nil {
			return nil, err
		}

		// Read the buffer to hash
		// TODO: stream it
		hashReq := &HashObjectRequest{
			hashObjectRequestInternal: internalReq,
		}
		hashReq.Content = reqBuf.Bytes()[(reqBuf.Len() - int(hashReq.Size)):]

		return hashReq, nil
	default:
		return nil, fmt.Errorf("unrecognized request '%s'", key)
	}
}
