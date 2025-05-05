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
		fmt.Println("Failed to read connection:", err)
		os.Exit(1)
	}
	request := strings.Split(string(buf), "\r\n")
	reqline := request[0]
	// reqheaders := request[1 : len(request)-2]
	// reqbody := request[len(request)-1]

	reqparts := strings.Split(reqline, " ")
	reqtarget := reqparts[1]
	// reqmethod := reqparts[0]
	// reqversion := reqparts[2]

	resversion := "HTTP/1.1"
	resstatuscode := "200"
	resstatusphrase := "OK"

	if reqtarget != "/" {
		resstatuscode = "400"
		resstatusphrase = "Not Found"
	}

	resstatus := strings.Join(
		[]string{
			resversion,
			resstatuscode,
			resstatusphrase,
		},
		" ")
	resheaders := ""
	resbody := ""
	response := strings.Join(
		[]string{
			resstatus,
			resheaders,
			resbody,
		},
		"\r\n")
	conn.Write([]byte(response))
	conn.Close()
}
