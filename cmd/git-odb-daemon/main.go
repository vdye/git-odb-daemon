package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/vdye/git-odb-daemon/internal/ipc"
	"github.com/vdye/git-odb-daemon/internal/storage"
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
	db := storage.NewFilesystemStorage(path)

	for {
		// Accept an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection received")

		// Handle the connection in a separate goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			for {
				req, err := ipc.ReadRequest(conn)
				if err != nil {
					fmt.Printf("error: failed to read request '%v'\n", err)
					ipc.WriteErrorResponse(conn)
					return
				}

				fmt.Printf("Request read: %s\n", req.Key())
				switch key := req.Key(); key {
				case "flush":
					continue
				case "EOF":
					return
				case "oid":
					oidReq, ok := req.(*ipc.GetOidRequest)
					if !ok {
						fmt.Printf("error: invalid request type for command '%s'\n", key)
						ipc.WriteErrorResponse(conn)
						return
					}

					objectInfo, err := db.EncodedObject(plumbing.AnyObject, oidReq.ObjectId.GitHash())
					if err != nil {
						fmt.Printf("error: failed to read object '%s'\n", oidReq.ObjectId.Hex())
						ipc.WriteErrorResponse(conn)
						return
					}

					// Populate the response
					resp := ipc.GetOidResponse{
						Key: [16]byte{'o', 'i', 'd'},
					}
					respOid, err := ipc.GitHashToObjectId(objectInfo.Hash())
					if err != nil {
						fmt.Printf("error: invalid hash '%s'\n", objectInfo.Hash())
						ipc.WriteErrorResponse(conn)
						return
					}
					resp.Oid = *respOid
					resp.Size = uint32(objectInfo.Size())
					resp.Type = int32(objectInfo.Type())
					// TODO: fill in more fields, stream object content if needed

					var contentReader io.ReadCloser
					if oidReq.WantContent != 0 {
						contentReader, err = objectInfo.Reader()
						if err != nil {
							fmt.Printf("error: failed to get object reader")
							ipc.WriteErrorResponse(conn)
							return
						}
					} else {
						contentReader = nil
					}

					err = resp.WriteResponse(conn, contentReader)
					if err != nil {
						fmt.Printf("error: failed to write response '%v'\n", err)
						return
					}
				default:
					fmt.Printf("error: unrecognized command '%s'\n", key)
					ipc.WriteErrorResponse(conn)
					return
				}
			}
		}(conn)
	}
}
