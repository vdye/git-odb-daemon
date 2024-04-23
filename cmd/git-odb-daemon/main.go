package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/vdye/git-odb-daemon/internal/db"
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

	// TODO: pick another database backend
	db, err := db.NewGitDb(path)
	if err != nil {
		log.Fatalf("unable to open DB connection: %v\n", err)
	}

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
				ipc.WriteErrorResponse(conn)
				return
			}

			fmt.Printf("Request read: %s\n", req.Key())
			switch key := req.Key(); key {
			case "oid":
				oidReq, ok := req.(*ipc.GetOidRequest)
				if !ok {
					fmt.Printf("error: invalid request type for command '%s'\n", key)
					ipc.WriteErrorResponse(conn)
					return
				}

				objectInfo, err := db.ReadObject(oidReq.ObjectId.GitHash(), false)
				if err != nil {
					fmt.Printf("error: failed to read object '%s'\n", oidReq.ObjectId.Hex())
					ipc.WriteErrorResponse(conn)
					return
				}
				fmt.Printf("size: %d\n", objectInfo.Size())
				fmt.Printf("object type: %s\n", objectInfo.Type().String())
			default:
				fmt.Printf("error: unrecognized command '%s'\n", key)
				ipc.WriteErrorResponse(conn)
				return
			}

			// TEMPORARY: return error until we have a DB to connect to
			ipc.WriteErrorResponse(conn)
		}(conn)
	}
}
