package test

import (
	"bytes"
	"encoding/binary"
	"testing"
)

type TestDataStruct struct {
	Int    int64
	Float  float64
	String string
}

func BenchmarkEndianness(b *testing.B) {
	buf := &bytes.Buffer{}
	b.Run("little endian", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := binary.Write(buf, binary.LittleEndian, TestDataStruct{42, 3.14, "test"}); err != nil {
				b.Fatalf("Unexpected error: '%+v'", err)
			}
		}
	})
	b.Run("big endian", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := binary.Write(buf, binary.BigEndian, TestDataStruct{42, 3.14, "test"}); err != nil {
				b.Fatalf("Unexpected error: '%+v'", err)
			}
		}
	})
}
