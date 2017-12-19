package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	GET      = "GET"
	QUIT     = "QUIT"
	SHUTDOWN = "SHUTDOWN"
)

var (
	file = flag.String("f", "test.txt", "file name to open")
)

type request struct {
	method string
	n      int
}

func ParseRequest(msg string) (*request, error) {
	parts := strings.Split(msg, " ")
	req := &request{method: parts[0]}

	switch req.method {
	case GET, QUIT, SHUTDOWN:
	default:
		return nil, errors.New("Unknown method")
	}

	switch len(parts) {
	case 0:
		return nil, errors.New("Invalid format")
	case 1:
		return req, nil
	case 2:
		_, err := fmt.Sscanf(parts[1], "%d", &req.n)
		if err != nil {
			return nil, err
		}
		return req, nil
	default:
		return nil, errors.New("Invalid format")
	}
}

func serveClient(conn net.Conn, file string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	l, err := NewLineMeta(f)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(conn)

	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			panic(err)
		}
		req, err := ParseRequest(strings.TrimSpace(string(line)))
		fmt.Printf("Got line '%s' %v\n", line, req)
		if err != nil {
			fmt.Printf("Err: %s\n", err)
			continue
		}
		switch req.method {
		case QUIT:
			err = conn.Close()
			if err != nil {
				fmt.Printf("Err closing: %s\n", err)
			}
			return
		case SHUTDOWN:
			os.Exit(0)
		case GET:
			buf, err := l.Line(req.n, f)
			if err != nil {
				panic(err)
			}
			conn.Write(append(buf, '\n'))
			fmt.Printf("Got %s\n", buf)
		}
	}
}

func main() {
	flag.Parse()
	f, err := os.Open(*file) // just test it opens before starting
	if err != nil {
		panic(err)
	}
	f.Close()

	l, err := net.Listen("tcp", "localhost:10497")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go serveClient(conn, *file)
	}
}
