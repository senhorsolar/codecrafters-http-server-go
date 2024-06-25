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

		go handle(conn)
	}
}

func handle(conn net.Conn) {

	defer conn.Close()

	reqBuf := make([]byte, 1024)
	n, err := conn.Read(reqBuf)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}
	fmt.Printf("Request: %s\n", reqBuf[:n])

	req := string(reqBuf[:n])
	parts := strings.Split(req, "\r\n")
	path := strings.Split(parts[0], " ")[1]

	switch {
	case path == "/":
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
	case strings.HasPrefix(path, "/echo/"):
		body := path[strings.LastIndex(path, "/")+1:]
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	case strings.HasPrefix(path, "/user-agent"):
		for _, part := range parts {
			if strings.HasPrefix(part, "User-Agent") {
				body := strings.Split(part, " ")[1]
				fmt.Printf("Body: %s\n", body)
				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
			}
		}
	default:
		fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
	}
}
