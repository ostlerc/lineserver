package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"io/ioutil"
	"math"
	"os"
)

type line struct {
	offset int64
	length int
}

type lineMeta struct {
	f *os.File
}

func (l *lineMeta) Line(line int, r io.ReaderAt) ([]byte, error) {
	var (
		offset int64
		length int32
	)

	if _, err := l.f.Seek(int64(line*12), 0); err != nil {
		return nil, err
	}
	if err := binary.Read(l.f, binary.BigEndian, &offset); err != nil {
		return nil, err
	}
	if err := binary.Read(l.f, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	buf := make([]byte, length, length)
	_, err := r.ReadAt(buf, offset)
	return buf, err
}

func (l *lineMeta) Close() error {
	l.f.Close()
	return os.Remove(l.f.Name())
}

func NewLineMeta(r io.Reader) (*lineMeta, error) {
	s := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, math.MaxInt32)

	f, err := ioutil.TempFile("", "lineserver")
	if err != nil {
		return nil, err
	}

	fileOffset := int64(0)
	for at := 0; s.Scan(); at++ {
		if err = binary.Write(f, binary.BigEndian, fileOffset); err != nil {
			return nil, err
		}
		if err = binary.Write(f, binary.BigEndian, int32(len(s.Bytes()))); err != nil {
			return nil, err
		}
		fileOffset += int64(len(s.Bytes())) + 1
	}

	return &lineMeta{f: f}, s.Err()
}
