package main

import (
	"fmt"
	"log"
	"net"

	"github.com/hconn7/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error listening on port %s", port)
	}

	defer listener.Close()
	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

	}

}
