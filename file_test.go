package main

import (
	"fmt"
	"os"
	"testing"
)

func TestLineFile(t *testing.T) {
	f, err := os.Open("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	l, err := NewLineMeta(f)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; ; i++ {
		v, ok := l.lineOffsets[i]
		if !ok {
			return
		}
		buf, err := l.Line(i, f)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("'%s' %v %v\n", string(buf), i, v)
	}
}
