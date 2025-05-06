package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

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
	req := parse_request(conn)
	res := handle_request(req)
	buf := serialize_response(res)

	_, err := conn.Write(buf)
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
	body    string
	headers map[string]string
}

func parse_request(conn net.Conn) request {
	reader := bufio.NewReader(conn)
	reqline, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read string:", err)
		os.Exit(1)
	}
	reqline = strings.Trim(reqline, "\r\n")
	reqparts := strings.Split(reqline, " ")
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read string:", err)
			os.Exit(1)
		}
		if line == "\r\n" {
			break
		}
		line = strings.Trim(line, "\r\n")
		headerparts := strings.Split(line, ": ")
		headers[headerparts[0]] = headerparts[1]
	}
	body := ""
	if val, ok := headers["Content-Length"]; ok {
		n, err := strconv.Atoi(val)
		if err != nil {
			fmt.Println("Failed to convert integer:", err)
			os.Exit(1)
		}
		buf := make([]byte, n)
		_, err = conn.Read(buf)
		if err != nil {
			fmt.Println("Failed to read from connection:", err)
			os.Exit(1)
		}
		body = string(buf)
	}
	return request{reqparts[0], reqparts[1], reqparts[2], body, headers}
}

type response struct {
	version string
	code    string
	phrase  string
	body    string
	headers []string
}

func handle_request(req request) response {
	res := response{"HTTP/1.1", "200", "OK", "", make([]string, 0)}
	if req.method == "GET" {
		handle_get_request(&req, &res)
	} else if req.method == "POST" {
		handle_post_request(&req, &res)
	}
	return res
}

func handle_get_request(req *request, res *response) {
	targetparts := strings.Split(req.target, "/")
	method := targetparts[1]
	args := make([]string, 0)
	if len(targetparts) > 2 {
		args = targetparts[2:]
	}
	if req.target == "/" {
		return
	}

	switch method {
	case "echo":
		if len(args) == 0 {
			res.code = "400"
			res.phrase = "Bad Request"
			return
		}
		msg := args[0]
		res.headers = append(res.headers, "Content-Type: text/plain")
		res.headers = append(res.headers, fmt.Sprintf("Content-Length: %d", len(msg)))
		res.headers = append(res.headers, "")
		res.body = msg
	case "user-agent":
		for _, header := range req.headers {
			headerparts := strings.Split(header, ": ")
			headertype := headerparts[0]
			headervalue := headerparts[1]
			if headertype == "User-Agent" {
				res.headers = append(res.headers, "Content-Type: text/plain")
				res.headers = append(res.headers, fmt.Sprintf("Content-Length: %d", len(headervalue)))
				res.headers = append(res.headers, "")
				res.body = headervalue
				break
			}
		}
	case "files":
		if len(args) == 0 {
			res.code = "400"
			res.phrase = "Bad Request"
			return
		}
		filename := args[0]
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			res.code = "400"
			res.phrase = "Not Found"
			return
		}
		buf, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("Failed to read file:", err)
			os.Exit(1)
		}
		res.headers = append(res.headers, "Content-Type: application/octet-stream")
		res.headers = append(res.headers, fmt.Sprintf("Content-Length: %d", len(buf)))
		res.headers = append(res.headers, "")
		res.body = string(buf)
	default:
		res.code = "400"
		res.phrase = "Not Found"
	}
}

func handle_post_request(req *request, res *response) {
	targetparts := strings.Split(req.target, "/")
	method := targetparts[1]
	args := make([]string, 0)
	if len(targetparts) > 2 {
		args = targetparts[2:]
	}

	switch method {
	case "files":
		if len(args) == 0 {
			res.code = "400"
			res.phrase = "Bad Request"
			return
		}
		filename := args[0]
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
		res.code = "201"
		res.phrase = "Created"
	default:
		res.code = "400"
		res.phrase = "Not Found"
	}
}

func serialize_response(res response) []byte {
	statusline := strings.Join([]string{res.version, res.code, res.phrase}, " ")
	headers := strings.Join(res.headers, "\r\n")
	return []byte(strings.Join([]string{statusline, headers, res.body}, "\r\n"))
}
