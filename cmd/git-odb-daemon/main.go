package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/vdye/git-odb-daemon/internal/ipc"
)

func main() {
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Connect to socket
	socket, err := net.Listen("unix", filepath.Join(path, "odb-over-ipc"))
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup the socket file.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(path)
		os.Exit(1)
	}()

	// TODO: connect to the DB

	for {
		// Accept an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a separate goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			req, err := ipc.ReadRequest(conn)
			if err != nil {
				conn.Write([]byte("error"))
				return
			}

			fmt.Printf("Request read: %s\n", req.Key())
			if req.Key() == "oid" {
				oidReq := req.(*ipc.GetOidRequest)
				if oidReq != nil {
					fmt.Printf("Requested OID: %s\n", oidReq.ObjectId.Hex())
					fmt.Printf("Flags: %s\n", strconv.FormatUint(uint64(oidReq.Flags), 2))
					fmt.Printf("WantContent: %d\n", oidReq.WantContent)
				}
			}

			// TEMPORARY: return error until we have a DB to connect to
			conn.Write([]byte("error"))
		}(conn)
	}
}
