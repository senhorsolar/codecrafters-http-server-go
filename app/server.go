package main

import (
	//"compress/gzip"
	"errors"
	"fmt"
	//"maps"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const CLRF = "\r\n"

type HttpRequest struct {
	method  string
	target  string
	version string
	headers map[string]string
	body    []byte
}

func parseHttpRequest(conn net.Conn) (*HttpRequest, error) {
	reqBuf := make([]byte, 1024)
	n, err := conn.Read(reqBuf)
	if err != nil {
		return nil, err
	}

	req := string(reqBuf[:n])
	parts := strings.Split(req, CLRF)

	reqLine := strings.Split(parts[0], " ")

	if len(reqLine) < 3 {
		return nil, errors.New("Invalid request line")
	}
	method := reqLine[0]
	target := reqLine[1]
	version := reqLine[2]

	headers := make(map[string]string)

	for _, part := range parts[1:(len(parts)-1)] { // all but first (req) and last (body)
		line := strings.Trim(part, " ")
		if len(line) > 0 {
			kv := strings.Split(part, ": ")
			if len(kv) < 2 {
				fmt.Printf("Error splitting %s into key: value pair\n", part)
				continue
			}
			headers[kv[0]] = kv[1]
		}
	}

	body := []byte(parts[len(parts)-1])

	return &HttpRequest{method: method, target: target, version: version, headers: headers, body: body}, nil
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

	httpRequest, err := parseHttpRequest(conn)
	if err != nil {
		fmt.Println("Error parsing http request: ", err.Error())
		errorResponse(conn)
		return
	}

	switch {
	case httpRequest.target == "/":
		okReponse(conn)
	case strings.HasPrefix(httpRequest.target, "/echo/"):
		body := strings.TrimPrefix(httpRequest.target, "/echo/")
		compressed := false

		val, ok := httpRequest.headers["Accept-Encoding"]
		if ok {
			for _, encType := range strings.Split(val, ", ") {
				if encType == "gzip" {
					compressed = true
				}
			}
		}
		if compressed {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		} else {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		}
	case strings.HasPrefix(httpRequest.target, "/user-agent"):
		body, ok := httpRequest.headers["User-Agent"]
		if ok {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		} else {
			errorResponse(conn)
		}
	case strings.HasPrefix(httpRequest.target, "/files/"):
		if (len(os.Args) < 3 || os.Args[1] != "--directory") {
			errorResponse(conn)
			return;
		}
		dir := os.Args[2]

		fileName := strings.Split(httpRequest.target, "files/")[1]
		fmt.Printf("file: %s\n", fileName)

		switch {
		case httpRequest.method == "GET":
			body, err := os.ReadFile(filepath.Join(dir, fileName))
			if err != nil {
				fmt.Println("Error reading file: ", err.Error())
				errorResponse(conn);
				return
			}
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		case httpRequest.method == "POST":
			body := httpRequest.body
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
