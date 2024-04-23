package ipc

import (
	"fmt"
	"net"
)

type GetOidResponse struct {
	Key
	Oid          ObjectId
	DeltaBaseOid ObjectId
	DiskSize     int64
	Size         uint32
	Whence       uint16
	Type         ObjectType
}

func WriteErrorResponse(conn net.Conn) {
	message := []byte("error")
	conn.Write([]byte(fmt.Sprintf("%04x", len(message)+4)))
	conn.Write(message)
	conn.Write([]byte("0000")) // flush
}
