package input

import (
	"fmt"
	"github.com/antchfx/xpath"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"testing"
)

func createXmlTestFile(t *testing.T, data string) (*os.File, error) {
	path := t.TempDir()
	fileName := "testOpen"

	file, err := os.Create(path + "/" + fileName)
	if err != nil {
		return nil, err
	}

	_, err = file.WriteString(data)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func TestXmlGet(t *testing.T) {
	file, err := createXmlTestFile(t, `
		<root>
			<item a="a0"><b>b0</b></item>
			<item a="a1"><b>b1</b></item>
			<item a="a2"><b>b2</b></item>
		</root>
	`)
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}
	defer file.Close()

	parsers := parser.List{"mock": parser.NewMock()}

	falseValue := false
	config := &config.XmlInput{
		Path:             file.Name(),
		DieOnInputChange: &falseValue,
		Logger:           logrus.NewEntry(logrus.StandardLogger()),
		RecordXPath:      "//item",
		Properties: []*config.XmlInputProperty{
			{
				Type:          config.XmlInputPropertyTypePrimitive,
				Name:          "a",
				Parser:        "mock",
				CompiledXPath: xpath.MustCompile("string(/@a)"),
			}, {
				Type:          config.XmlInputPropertyTypePrimitive,
				Name:          "b",
				Parser:        "mock",
				CompiledXPath: xpath.MustCompile("string(/b)"),
			},
		},
	}

	xml, err := NewXml(config, parsers)
	if err != nil {
		t.Error(err)
	}

	t.Run("normal", func(t *testing.T) {
		// Testing a normal read
		row, err := xml.Get(13)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		expect := "a0"
		if result, _ := row.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
		expect = "b0"
		if result, _ := row.Get("b"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}

		// Testing if the position in the file and buffer are properly set
		// when it has already been used once
		row, err = xml.Get(46)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		expect = "a1"
		if result, _ := row.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
		expect = "b1"
		if result, _ := row.Get("b"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
	})
	t.Run("parallel", func(t *testing.T) {
		// Executing two read operations in parallel to test the safety
		errorMessages1 := make(chan string)
		errorMessages2 := make(chan string)

		go (func() {
			expect := "a1"
			row, err := xml.Get(46)
			if err != nil {
				errorMessages1 <- fmt.Sprintf("Expected no error, got '%v'", err)
			}
			if result, _ := row.Get("a"); result != expect {
				errorMessages1 <- fmt.Sprintf("Expected '%v', got '%v'", expect, result)
			}
			close(errorMessages1)
		})()
		go (func() {
			expect := "a0"
			row, err := xml.Get(13)
			if err != nil {
				errorMessages2 <- fmt.Sprintf("Expected no error, got '%v'", err)
			}
			if result, _ := row.Get("a"); result != expect {
				errorMessages2 <- fmt.Sprintf("Expected '%v', got '%v'", expect, result)
			}
			close(errorMessages2)
		})()

		for {
			select {
			case errorMessage, ok := <-errorMessages1:
				if ok {
					t.Fatal(errorMessage)
				} else {
					errorMessages1 = nil
				}
			case errorMessage, ok := <-errorMessages2:
				if ok {
					t.Fatal(errorMessage)
				} else {
					errorMessages2 = nil
				}
			}

			if errorMessages1 == nil && errorMessages2 == nil {
				break
			}
		}
	})
	t.Run("from IterateAll", func(t *testing.T) {
		iterator, end, err := xml.IterateAll()
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		defer func() {
			err := end()
			if err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
		}()

		index := 0
		var secondRow record.Record = nil
		for {
			record, err := iterator()
			if err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
			if record == nil {
				break
			}

			if index == 1 {
				secondRow = record
			}
			index++
		}
		if secondRow == nil {
			t.Fatalf("Expected a record, got '%v'", secondRow)
		}

		record, err := xml.Get(secondRow.Position())
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		expect := "a1"
		if result, _ := secondRow.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
		if result, _ := record.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
	})
}

func TestXmlSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, err := createXmlTestFile(t, data)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		falseValue := false
		config := &config.XmlInput{
			Path:             file.Name(),
			DieOnInputChange: &falseValue,
			RecordXPath:      "test",
			Logger:           logrus.NewEntry(logrus.StandardLogger()),
			Properties:       []*config.XmlInputProperty{},
		}

		xml, err := NewXml(config, parsers)
		if err != nil {
			t.Error(err)
		}

		size, err := xml.Size()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Fatalf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestXmlIterateAll(t *testing.T) {
	testCases := []struct {
		name              string
		file              string
		expectedRows      [][]interface{}
		expectedPositions []int64
	}{
		{
			name:              "normal",
			file:              `<root><item a="a0"><b>b0</b></item><item a="a1"><b>b1</b></item></root>`,
			expectedRows:      [][]interface{}{{"a0", "b0"}, {"a1", "b1"}},
			expectedPositions: []int64{0, 35},
		}, {
			name:              "empty",
			file:              `<root></root>`,
			expectedRows:      [][]interface{}{},
			expectedPositions: []int64{},
		}, {
			name:              "wrong tag",
			file:              `<root><item a="a0"><b>b0</b></item><non-item><foo>bar</foo></non-item><item a="a1"><b>b1</b></item></root>`,
			expectedRows:      [][]interface{}{{"a0", "b0"}, {"a1", "b1"}},
			expectedPositions: []int64{0, 35},
		}, {
			name:              "end after one row",
			file:              `<root><item a="a0"><b>b0</b></item></root>`,
			expectedRows:      [][]interface{}{{"a0", "b0"}},
			expectedPositions: []int64{0},
		}, {
			name:              "no match for the xpath",
			file:              `<root><item a="a0"><c>c0</c></item></root>`,
			expectedRows:      [][]interface{}{{"a0", ""}}, // Xpath does not return nil, but empty string
			expectedPositions: []int64{0},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := createXmlTestFile(t, testCase.file)
			if err != nil {
				t.Fatalf("Unexpected error: '%+v'", err)
			}
			defer file.Close()

			parsers := parser.List{"mock": parser.NewMock()}

			falseValue := false
			config := &config.XmlInput{
				Path:             file.Name(),
				DieOnInputChange: &falseValue,
				RecordXPath:      "//item",
				Logger:           logrus.NewEntry(logrus.StandardLogger()),
				Properties: []*config.XmlInputProperty{
					{
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          "a",
						Parser:        "mock",
						CompiledXPath: xpath.MustCompile("string(/@a)"),
					}, {
						Type:          config.XmlInputPropertyTypePrimitive,
						Name:          "b",
						Parser:        "mock",
						CompiledXPath: xpath.MustCompile("string(/b)"),
					},
				},
			}
			xml, err := NewXml(config, parsers)
			if err != nil {
				t.Error(err)
			}

			iterator, end, err := xml.IterateAll()
			if err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
			defer func() {
				err := end()
				if err != nil {
					t.Fatalf("Expected no error, got '%v'", err)
				}
			}()

			for i := 0; i < len(testCase.expectedRows); i++ {
				if record, err := iterator(); err != nil {
					t.Error(err)
				} else {
					if record == nil {
						break
					}

					for j, k := range []string{"a", "b"} {
						result, err := record.Get(k)
						if err != nil {
							t.Fatalf("Got error '%v', expected '%v' for record %v, property number %v", err, testCase.expectedRows[i][j], i, j)
							continue
						}

						expected := testCase.expectedRows[i][j]
						if !(result == nil && expected == nil) && result != expected {
							t.Fatalf("Received '%v', expected '%v' for record %v, property number %v", result, testCase.expectedRows[i][j], i, j)
						}
					}

					if got, expect := record.Position(), testCase.expectedPositions[i]; got != expect {
						t.Fatalf("Got position '%v', expected '%v' for record %v", got, expect, i)
					}
				}

				// Asserts that IterateAll does not fail with concurrent accesses
				xml.reader.Seek(0, io.SeekStart)
			}
		})
	}
}
