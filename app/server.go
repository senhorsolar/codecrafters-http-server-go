package main

import (
	"fmt"
	// Uncomment this block to pass the first stage
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		req := make([]byte, 1024);
		conn.Read(req);
		if strings.HasPrefix(string(req), "GET / HTTP/1.1") {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n");
		}
		else {
			fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n");
		}
	}
}
