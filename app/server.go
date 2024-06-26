package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"path/filepath"
)

type RequestLine struct {
	method  string
	target  string
	version string
}

type Headers struct {
	host string
	port int
	userAgent string
	accept string
	contentType string
	contentLength int
}

type HttpRequest struct {
	request RequestLine
	headers Headers
	body string
}

//func parseHttpRequest()

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
	reqLine := strings.Split(parts[0], " ")
	path := reqLine[1];

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

		fileName := strings.Split(path, "files/")[1]
		fmt.Printf("file: %s\n", fileName)

		switch {
		case reqLine[0] == "GET":
			body, err := os.ReadFile(filepath.Join(dir, fileName))
			if err != nil {
				fmt.Println("Error reading file: ", err.Error())
				errorResponse(conn);
				return
			}
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		case reqLine[0] == "POST":
			body := parts[len(parts)-1]
			err := os.WriteFile(filepath.Join(dir, fileName), []byte(body), 0644)
			if err != nil {
				fmt.Println("Error creating file: ", err.Error())
				errorResponse(conn)
				return
			}
			fmt.Fprintf(conn, "HTTP/1.1 201 Created\r\n\r\n")
		default:
			errorResponse(conn)
		}
	default:
		errorResponse(conn)
	}
}
