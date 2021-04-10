package input

import (
	"encoding/xml"
	"github.com/antchfx/xpath"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"rods/pkg/config"
	"rods/pkg/parser"
	"rods/pkg/record"
	"sync"
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

func TestXmlHasColumn(t *testing.T) {
	file, err := createXmlTestFile(t, "")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	defer file.Close()

	parsers := parser.List{"mock": parser.NewMock()}

	falseValue := false
	config := &config.XmlInput{
		Path:             file.Name(),
		DieOnInputChange: &falseValue,
		RecordXPath:      "test",
		Logger:           logrus.NewEntry(logrus.StandardLogger()),
		Columns: []*config.XmlInputColumn{
			{Name: "a", Parser: "mock"},
			{Name: "b", Parser: "mock"},
		},
		ColumnIndexByName: map[string]int{
			"a": 0,
			"b": 1,
		},
	}

	xml, err := NewXml(config, parsers)
	if err != nil {
		t.Error(err)
	}

	t.Run("true", func(t *testing.T) {
		if !xml.HasColumn("a") {
			t.Errorf("Expected to have column 'a', got false")
		}
		if !xml.HasColumn("b") {
			t.Errorf("Expected to have column 'b', got false")
		}
	})
	t.Run("false", func(t *testing.T) {
		if xml.HasColumn("wrong") {
			t.Errorf("Expected to not have column 'wrong', got true")
		}
	})
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
		t.Errorf("Unexpected error: '%+v'", err)
	}
	defer file.Close()

	parsers := parser.List{"mock": parser.NewMock()}

	falseValue := false
	config := &config.XmlInput{
		Path:             file.Name(),
		DieOnInputChange: &falseValue,
		Logger:           logrus.NewEntry(logrus.StandardLogger()),
		RecordXPath:      "//item",
		Columns: []*config.XmlInputColumn{
			{
				Name:          "a",
				Parser:        "mock",
				CompiledXPath: xpath.MustCompile("string(/@a)"),
			}, {
				Name:          "b",
				Parser:        "mock",
				CompiledXPath: xpath.MustCompile("string(/b)"),
			},
		},
		ColumnIndexByName: map[string]int{
			"a": 0,
			"b": 1,
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
			t.Errorf("Expected no error, got '%v'", err)
		}
		expect := "a0"
		if result, _ := row.Get("a"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}
		expect = "b0"
		if result, _ := row.Get("b"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}

		// Testing if the position in the file and buffer are properly set
		// when it has already been used once
		row, err = xml.Get(46)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		expect = "a1"
		if result, _ := row.Get("a"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}
		expect = "b1"
		if result, _ := row.Get("b"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}
	})
	t.Run("parallel", func(t *testing.T) {
		// Executing two read operations in parallel to test the safety
		wait := sync.WaitGroup{}
		wait.Add(2)
		go (func() {
			expect := "a1"
			row, err := xml.Get(46)
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
			if result, _ := row.Get("a"); result != expect {
				t.Errorf("Expected '%v', got '%v'", expect, result)
			}
			wait.Done()
		})()
		go (func() {
			expect := "a0"
			row, err := xml.Get(13)
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
			if result, _ := row.Get("a"); result != expect {
				t.Errorf("Expected '%v', got '%v'", expect, result)
			}
			wait.Done()
		})()
		wait.Wait()
	})
	t.Run("from IterateAll", func(t *testing.T) {
		index := 0
		var secondRow record.Record = nil
		for result := range xml.IterateAll() {
			if result.Error != nil {
				t.Errorf("Expected no error, got '%v'", result.Error)
			}
			if index == 1 {
				secondRow = result.Record
			}
			index++
		}
		if secondRow == nil {
			t.Errorf("Expected a record, got '%v'", secondRow)
		}

		record, err := xml.Get(secondRow.Position())
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}

		expect := "a1"
		if result, _ := secondRow.Get("a"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}
		if result, _ := record.Get("a"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}
	})
}

func TestXmlSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, err := createXmlTestFile(t, data)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		falseValue := false
		config := &config.XmlInput{
			Path:              file.Name(),
			DieOnInputChange:  &falseValue,
			RecordXPath:       "test",
			Logger:            logrus.NewEntry(logrus.StandardLogger()),
			Columns:           []*config.XmlInputColumn{},
			ColumnIndexByName: map[string]int{},
		}

		xml, err := NewXml(config, parsers)
		if err != nil {
			t.Error(err)
		}

		size, err := xml.Size()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Errorf("Expected to get a size of '%v', got '%+v'", len(data), size)
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
				t.Errorf("Unexpected error: '%+v'", err)
			}
			defer file.Close()

			parsers := parser.List{"mock": parser.NewMock()}

			falseValue := false
			config := &config.XmlInput{
				Path:             file.Name(),
				DieOnInputChange: &falseValue,
				RecordXPath:      "//item",
				Logger:           logrus.NewEntry(logrus.StandardLogger()),
				Columns: []*config.XmlInputColumn{
					{
						Name:          "a",
						Parser:        "mock",
						CompiledXPath: xpath.MustCompile("string(/@a)"),
					}, {
						Name:          "b",
						Parser:        "mock",
						CompiledXPath: xpath.MustCompile("string(/b)"),
					},
				},
				ColumnIndexByName: map[string]int{
					"a": 0,
					"b": 1,
				},
			}
			xml, err := NewXml(config, parsers)
			if err != nil {
				t.Error(err)
			}

			resultsChannel := xml.IterateAll()
			for i := 0; i < len(testCase.expectedRows); i++ {
				if result := <-resultsChannel; result.Error != nil {
					t.Error(err)
				} else {
					for j, k := range []string{"a", "b"} {
						result, err := result.Record.Get(k)
						if err != nil {
							t.Errorf("Got error '%v', expected '%v' for record %v, column number %v", err, testCase.expectedRows[i][j], i, j)
							continue
						}

						expected := testCase.expectedRows[i][j]
						if !(result == nil && expected == nil) && result != expected {
							t.Errorf("Received '%v', expected '%v' for record %v, column number %v", result, testCase.expectedRows[i][j], i, j)
						}
					}

					if got, expect := result.Record.Position(), testCase.expectedPositions[i]; got != expect {
						t.Errorf("Got position '%v', expected '%v' for record %v", got, expect, i)
					}
				}

				// Asserts that IterateAll does not fail with concurrent accesses
				xml.reader.Seek(0, io.SeekStart)
			}
		})
	}
}

func TestXmlGetOuterXml(t *testing.T) {
	parsers := parser.List{"string": parser.NewMock()}
	falseValue := false
	config := &config.XmlInput{
		DieOnInputChange: &falseValue,
		RecordXPath:      "test",
		Logger:           logrus.NewEntry(logrus.StandardLogger()),
	}

	t.Run("normal", func(t *testing.T) {
		xmlData := "<root attribute=\"42\"><a>a</a><b>b</b></root>"
		file, err := createXmlTestFile(t, xmlData)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		xmlInput, err := NewXml(config, parsers)
		if err != nil {
			t.Error(err)
		}

		token, err := xmlInput.xmlDecoder.Token()
		if err != nil {
			t.Error(err)
		}

		got, err := xmlInput.getOuterXml(xmlInput.xmlDecoder, token.(xml.StartElement))
		if err != nil {
			t.Error(err)
		}
		if string(got) != xmlData {
			t.Errorf("Expected to get xml '%v', got '%v'", xmlData, got)
		}
	})
}
