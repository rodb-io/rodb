package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/antchfx/xmlquery"
	"io"
	"reflect"
	"strings"
	"unsafe"
)

func main() {
	reader := strings.NewReader(`
		<list>
			<item><x><a>a0</a><b>b0</b></x></item>
			<item><x><a>a1</a><b>b1</b></x></item>
			<item><x><a>a2</a><b>b2</b></x></item>
			<item><x><a>a3</a><b>b3</b></x></item>
			<item><x><a>a4</a><b>b4</b></x></item>
			<item><x><a>a5</a><b>b5</b></x></item>
			<item><x><a>a6</a><b>b6</b></x></item>
			<item><x><a>a7</a><b>b7</b></x></item>
			<item><x><a>a8</a><b>b8</b></x></item>
			<item><x><a>a9</a><b>b9</b></x></item>
		</list>
	`)

	stream, err := xmlquery.CreateStreamParser(reader, "/list/item")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	decoder := getDecoder(stream)
	buffer := getBuffer(stream)

	for {
		fmt.Printf("Position: %v\n", decoder.InputOffset())

		_, err := reader.Seek(387, io.SeekStart)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}
		buffer.Reset(reader)
		resetCache(stream)

		node, err := stream.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}

		fmt.Printf("Node: %#v\n", node)
	}
}

func getDecoder(parser *xmlquery.StreamParser) *xml.Decoder {
	p := reflect.ValueOf(parser).Elem().FieldByName("p")
	decoder := p.Elem().FieldByName("decoder")
	decoderInterface := reflect.NewAt(
		decoder.Type(),
		unsafe.Pointer(decoder.UnsafeAddr()),
	).Elem().Interface()

	return decoderInterface.(*xml.Decoder)
}

func getBuffer(parser *xmlquery.StreamParser) *bufio.Reader {
	p := reflect.ValueOf(parser).Elem().FieldByName("p")
	reader := p.Elem().FieldByName("reader")
	buffer := reader.Elem().FieldByName("buffer")

	bufferInterface := reflect.NewAt(
		buffer.Type(),
		unsafe.Pointer(buffer.UnsafeAddr()),
	).Elem().Interface()

	return bufferInterface.(*bufio.Reader)
}

func resetCache(parser *xmlquery.StreamParser) {
	p := reflect.ValueOf(parser).Elem().FieldByName("p")
	decoder := p.Elem().FieldByName("reader")

	cacheLen := decoder.Elem().FieldByName("cacheLen")

	reflect.NewAt(
		cacheLen.Type(),
		unsafe.Pointer(cacheLen.UnsafeAddr()),
	).Elem().Set(reflect.ValueOf(0))
}
