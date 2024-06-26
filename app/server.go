package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"path/filepath"
)

func main() {

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

func okReponse(conn net.Conn) {
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
}

func errorResponse(conn net.Conn) {
	fmt.Fprintf(conn, "HTTP/1.1 404 Not Found\r\n\r\n")
}

func handle(conn net.Conn) {

	defer conn.Close()

	reqBuf := make([]byte, 1024)
	n, err := conn.Read(reqBuf)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		errorResponse(conn)
		return
	}
	fmt.Printf("Request: %s\n", reqBuf[:n])

	req := string(reqBuf[:n])
	parts := strings.Split(req, "\r\n")
	path := strings.Split(parts[0], " ")[1]

	switch {
	case path == "/":
		okReponse(conn)
	case strings.HasPrefix(path, "/echo/"):
		body := path[strings.LastIndex(path, "/")+1:]
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	case strings.HasPrefix(path, "/user-agent"):
		ok := false
		for _, part := range parts {
			if strings.HasPrefix(part, "User-Agent") {
				body := strings.Split(part, " ")[1]
				fmt.Printf("Body: %s\n", body)
				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
			}
		}
		if (ok == false) {
			errorResponse(conn)
		}
	case strings.HasPrefix(path, "/files/"):
		if (len(os.Args) < 3 || os.Args[1] != "--directory") {
			errorResponse(conn)
			return;
		}
		dir := os.Args[2]
		file := strings.Split(path, "files/")[1]
		fmt.Printf("file: %s\n", file)
		body, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			fmt.Println ("Error reading file: ", err.Error())
			errorResponse(conn);
			return
		}
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	default:
		errorResponse(conn)
	}
}
