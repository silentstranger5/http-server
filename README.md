# HTTP server in Go

This is an HTTP server written in Go. It is built on top of TCP/IP stack.

## Features

- Concurrent connections
- Echo (`GET /echo/{message}`)
- User Agent (`GET /user-agent`)
- Read file (`GET /files/{filename}`)
- Send file (`POST /files/{filename}` with the file body)

## How to build

```bash
git clone https://github.com/silentstranger5/http-server
cd http-server
go build ./app
./app
```

You can specify filesystem directory with the special flag:

```bash
./app --directory /tmp
```
