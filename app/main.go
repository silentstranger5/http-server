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

	handle_conn(conn)
}

func handle_conn(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read connection:", err)
		os.Exit(1)
	}

	req := parse_request(buf)
	res := handle_request(req)
	buf = serialize_response(res)

	_, err = conn.Write(buf)
	if err != nil {
		fmt.Println("Failed to write to connection:", err)
		os.Exit(1)
	}
	conn.Close()
}

type request struct {
	method  string
	target  string
	version string
	headers []string
	body    string
}

func parse_request(buf []byte) request {
	req := strings.Split(string(buf), "\r\n")
	reqline := req[0]
	reqheaders := req[1 : len(req)-2]
	reqbody := req[len(req)-1]

	reqparts := strings.Split(reqline, " ")
	reqmethod := reqparts[0]
	reqtarget := reqparts[1]
	reqversion := reqparts[2]

	return request{
		method:  reqmethod,
		target:  reqtarget,
		version: reqversion,
		headers: reqheaders,
		body:    reqbody,
	}
}

type response struct {
	version string
	code    string
	phrase  string
	body    string
	headers []string
}

func handle_request(req request) response {
	resversion := "HTTP/1.1"
	rescode := "200"
	resphrase := "OK"
	resheaders := make([]string, 0)
	resbody := ""

	if req.target == "/" {
		return response{resversion, rescode, resphrase, resbody, resheaders}
	}

	targetparts := strings.Split(req.target, "/")
	if len(targetparts) < 3 {
		rescode = "400"
		resphrase = "Not Found"
		return response{resversion, rescode, resphrase, resbody, resheaders}
	}

	method := targetparts[1]
	switch method {
	case "echo":
		msg := targetparts[2]
		resheaders = append(resheaders, "Content-Type: text/plain")
		resheaders = append(resheaders, fmt.Sprintf("Content-Length: %d", len(msg)))
		resheaders = append(resheaders, "")
		resbody = msg
	default:
		rescode = "400"
		resphrase = "Not Found"
	}

	return response{resversion, rescode, resphrase, resbody, resheaders}
}

func serialize_response(res response) []byte {
	statusline := strings.Join([]string{res.version, res.code, res.phrase}, " ")
	headers := strings.Join(res.headers, "\r\n")
	return []byte(strings.Join([]string{statusline, headers, res.body}, "\r\n"))
}
