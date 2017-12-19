package main

import (
	"bufio"
	"io"
	"math"
)

type line struct {
	offset int64
	length int
}

type lineMeta struct {
	// lineOffsets is a map of line number to file offset/byte count
	lineOffsets map[int]*line
}

func (l *lineMeta) Line(line int, r io.ReaderAt) ([]byte, error) {
	v, ok := l.lineOffsets[line]
	if !ok {
		return nil, io.EOF
	}
	buf := make([]byte, v.length, v.length)
	_, err := r.ReadAt(buf, v.offset)
	return buf, err
}

func NewLineMeta(r io.Reader) (*lineMeta, error) {
	s := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, math.MaxInt32)
	offsets := map[int]*line{}

	fileOffset := int64(0)
	for at := 0; s.Scan(); at++ {
		offsets[at] = &line{offset: fileOffset, length: len(s.Bytes())}
		fileOffset += int64(len(s.Bytes())) + 1
	}

	return &lineMeta{lineOffsets: offsets}, s.Err()
}
