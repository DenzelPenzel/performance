// go test -bench . -benchmem -memprofile p.out -gcflags -m=2
// go test -bench . -benchtime 3s -benchmem -memprofile p.out

// go tool pprof -noinlines p.out

package main

import (
	"bytes"
	"testing"
)

var output bytes.Buffer
var in = assembleInputStream()
var find = []byte("elvis")
var repl = []byte("Elvis")

// Capture the time it takes to execute algorithm
func BenchmarkBase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		output.Reset()
		match(in, find, repl, &output)
	}
}

func BenchmarkZeroAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		output.Reset()
		matchFinal(in, find, repl, &output)
	}
}
