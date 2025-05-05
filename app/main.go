package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	handleConn(conn)
}

func handleConn(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read connection")
		os.Exit(1)
	}
	version := "HTTP/1.1"
	statuscode := "200"
	statusphrase := "OK"
	status := strings.Join(
		[]string{
			version,
			statuscode,
			statusphrase,
		},
		" ")
	headers := ""
	body := ""
	msg := strings.Join(
		[]string{
			status,
			headers,
			body,
		},
		"\r\n")
	conn.Write([]byte(msg))
	conn.Close()
}
