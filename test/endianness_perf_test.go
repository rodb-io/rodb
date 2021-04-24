package test

import (
	"bytes"
	"encoding/binary"
	"testing"
)

const iteratorBenchmarkIterations = 1000000

type TestDataStruct struct {
	Int    int64
	Float  float64
	String string
}

func BenchmarkEndianness(b *testing.B) {
	buf := &bytes.Buffer{}
	b.Run("little endian", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			binary.Write(buf, binary.LittleEndian, TestDataStruct{42, 3.14, "test"})
		}
	})
	b.Run("big endian", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			binary.Write(buf, binary.BigEndian, TestDataStruct{42, 3.14, "test"})
		}
	})
}
