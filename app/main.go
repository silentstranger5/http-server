package main

import (
	"flag"
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

	dir := flag.String("directory", ".", "Specify directory to use")
	flag.Parse()

	_, err = os.Stat(*dir)
	if os.IsNotExist(err) {
		fmt.Println("Directory", *dir, "does not exist")
		os.Exit(1)
	}

	err = os.Chdir(*dir)
	if err != nil {
		fmt.Println("Failed to change directory:", err)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handle_conn(conn)
	}
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

	if req.method == "GET" {
		targetparts := strings.Split(req.target, "/")
		method := targetparts[1]
		switch method {
		case "echo":
			msg := targetparts[2]
			resheaders = append(resheaders, "Content-Type: text/plain")
			resheaders = append(resheaders, fmt.Sprintf("Content-Length: %d", len(msg)))
			resheaders = append(resheaders, "")
			resbody = msg
		case "user-agent":
			for _, header := range req.headers {
				headerparts := strings.Split(header, ": ")
				headertype := headerparts[0]
				headervalue := headerparts[1]
				if headertype == "User-Agent" {
					resheaders = append(resheaders, "Content-Type: text/plain")
					resheaders = append(resheaders, fmt.Sprintf("Content-Length: %d", len(headervalue)))
					resheaders = append(resheaders, "")
					resbody = headervalue
					break
				}
			}
		case "files":
			filename := targetparts[2]
			_, err := os.Stat(filename)
			if os.IsNotExist(err) {
				rescode = "400"
				resphrase = "Not Found"
				return response{resversion, rescode, resphrase, resbody, resheaders}
			}
			buf, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println("Failed to read file:", err)
				os.Exit(1)
			}
			resheaders = append(resheaders, "Content-Type: application/octet-stream")
			resheaders = append(resheaders, fmt.Sprintf("Content-Length: %d", len(buf)))
			resheaders = append(resheaders, "")
			resbody = string(buf)
		default:
			rescode = "400"
			resphrase = "Not Found"
		}
	}

	if req.method == "POST" {
		targetparts := strings.Split(req.target, "/")
		method := targetparts[1]
		switch method {
		case "files":
			filename := targetparts[2]
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Failed to create file:", err)
				os.Exit(1)
			}
			defer file.Close()
			_, err = file.WriteString(req.body)
			if err != nil {
				fmt.Println("Failed to write to file", filename, ":", err)
				os.Exit(1)
			}
			rescode = "201"
			resphrase = "Created"
		default:
			rescode = "400"
			resphrase = "Not Found"
		}
	}

	return response{resversion, rescode, resphrase, resbody, resheaders}
}

func serialize_response(res response) []byte {
	statusline := strings.Join([]string{res.version, res.code, res.phrase}, " ")
	headers := strings.Join(res.headers, "\r\n")
	return []byte(strings.Join([]string{statusline, headers, res.body}, "\r\n"))
}
