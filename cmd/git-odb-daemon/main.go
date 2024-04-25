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
	socketPath := filepath.Join(path, "odb-over-ipc")
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: pick another database backend
	db := storage.NewFilesystemStorage(path)

	// Cleanup the socket file.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		db.Close()
		os.Remove(socketPath)
		os.Exit(1)
	}()

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
				case "hash-object":
					hashReq, ok := req.(*ipc.HashObjectRequest)
					if !ok {
						fmt.Printf("error: invalid request type for command '%s'\n", key)
						ipc.WriteErrorResponse(conn)
						return
					}

					obj := db.NewEncodedObject()
					obj.SetType(plumbing.ObjectType(hashReq.Type))
					obj.SetSize(int64(hashReq.Size))
					objWriter, err := obj.Writer()
					if err != nil {
						fmt.Printf("error: could not write object\n")
						ipc.WriteErrorResponse(conn)
						return
					}

					n, err := objWriter.Write(hashReq.Content)
					if err != nil {
						fmt.Printf("error: failed to read object contents: %v\n", err)
						ipc.WriteErrorResponse(conn)
						return
					} else if n != int(hashReq.Size) {
						fmt.Printf("error: incorrect write size (expected %d, got %d)\n", hashReq.Size, n)
						ipc.WriteErrorResponse(conn)
						return
					}

					var oid plumbing.Hash
					if hashReq.Flags&1 == 0 {
						// Just hash, don't write to ODB
						oid = obj.Hash()
					} else {
						// Write the object to storage
						oid, err = db.SetEncodedObject(obj)
						if err != nil {
							fmt.Printf("error: failed to write object: %v\n", err)
							ipc.WriteErrorResponse(conn)
							return
						}
					}

					if oid.IsZero() {
						fmt.Printf("error: could not compute hash\n")
						ipc.WriteErrorResponse(conn)
						return
					}

					// Populate the response
					resp := ipc.HashObjectResponse{}
					copy(resp.Key[:len(key)], key)

					respOid, err := ipc.GitHashToObjectId(oid)
					if err != nil {
						fmt.Printf("error: invalid hash '%s'\n", oid)
						ipc.WriteErrorResponse(conn)
						return
					}
					resp.Oid = *respOid

					err = resp.WriteResponse(conn)
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
