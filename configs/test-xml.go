package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
)

func main() {
	reader := strings.NewReader(`
		<list>
			<item test="42"><x><a>a0</a><b>b0</b></x></item>
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
	d := xml.NewDecoder(reader)

	// _, err := reader.Seek(387, io.SeekStart)
	// if err != nil {
	// 	log.Fatalf("Error: %s", err)
	// }

	position := int64(0)
	for {
		token, err := d.Token()
		if token == nil || err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error decoding token: %s", err)
		}

		element, isStartElement := token.(xml.StartElement)
		if isStartElement {
			if element.Name.Local == "item" {
				result := struct {
					XMLName xml.Name
					Attr []xml.Attr  `xml:",any,attr"`
					InnerXML []byte `xml:",innerxml"`
				}{}
				err := d.DecodeElement(&result, &element)
				if err != nil {
					log.Fatalf("Error: %s", err)
				}

				test, err := xml.Marshal(result)
				if err != nil {
					log.Fatalf("Error: %s", err)
				}

				fmt.Printf("Position %v, item: %v\n", position, string(test))

				// The position must be updated here, because the one that allows
				// retrieving an item is the one after the closing tag of the
				// previous item
				position = d.InputOffset()
			}
		}
	}
}
