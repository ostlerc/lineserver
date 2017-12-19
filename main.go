package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
)

const (
	GET      = "GET"
	QUIT     = "QUIT"
	SHUTDOWN = "SHUTDOWN"
)

var (
	addr = flag.String("addr", "localhost:10497", "server listen address")
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

func serveClient(l *lineMeta, conn net.Conn, file string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(conn)

	OK := []byte("OK\r\n")
	ERR := []byte("ERR\r\n")

	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			fmt.Printf("Error reading slice %v\n", err)
			conn.Write(ERR)
			err = conn.Close()
			if err != nil {
				fmt.Printf("Err closing: %s\n", err)
			}
			return
		}
		req, err := ParseRequest(strings.TrimSpace(string(line)))
		if err != nil {
			fmt.Printf("Err: %s\n", err)
			conn.Write(ERR)
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
			l.Close()
			os.Exit(0)
		case GET:
			buf, err := l.Line(req.n-1, f)
			if err != nil {
				fmt.Println("Err", err)
				conn.Write(ERR)
			} else {
				conn.Write(OK)
				conn.Write(append(buf, '\n'))
			}
		}
	}
}

func main() {
	flag.Parse()
	f, err := os.Open(*file) // test file opens fine and also populate meta data
	if err != nil {
		panic(err)
	}
	meta, err := NewLineMeta(f)
	if err != nil {
		panic(err)
	}
	f.Close()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		meta.Close()
		os.Exit(0)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go serveClient(meta, conn, *file)
	}

}
