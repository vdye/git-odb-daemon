package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type IpcRequest interface {
	Key() string
}

type GetOidRequest struct {
	ObjectId
	Flags       uint16
	WantContent byte
}

func (r *GetOidRequest) Key() string {
	return "oid"
}

func ReadRequest(conn net.Conn) (IpcRequest, error) {
	// First, read the size of the request
	sizeStr := make([]byte, 4)
	n, err := conn.Read(sizeStr)
	if err != nil {
		return nil, err
	} else if n < 4 {
		return nil, fmt.Errorf("could not read size")
	}

	reqSize, err := strconv.ParseUint(string(sizeStr), 16, 32)
	if err != nil {
		return nil, err
	}

	reqBuf := bytes.NewBuffer(make([]byte, reqSize))
	n, err = conn.Read(reqBuf.Bytes())
	if err != nil {
		return nil, err
	} else if n < int(reqSize) {
		return nil, fmt.Errorf("could not read request")
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
	default:
		return nil, fmt.Errorf("unrecognized request '%s'", key)
	}
}
