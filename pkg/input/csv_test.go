package input

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"github.com/rodb-io/rodb/pkg/input/record"
	"github.com/rodb-io/rodb/pkg/parser"
	"testing"
)

func createCsvTestFile(t *testing.T, data string) (*os.File, error) {
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

func TestCsvGet(t *testing.T) {
	file, err := createCsvTestFile(t, "test1,test2\n\ntest3")
	if err != nil {
		t.Fatalf("Unexpected error: '%+v'", err)
	}
	defer file.Close()

	parsers := parser.List{"mock": parser.NewMock()}

	logger := logrus.StandardLogger()
	logger.Level = logrus.ErrorLevel

	falseValue := false
	config := &CsvConfig{
		Path:             file.Name(),
		IgnoreFirstRow:   false,
		DieOnInputChange: &falseValue,
		Delimiter:        ",",
		Logger:           logrus.NewEntry(logger),
		Columns: []*CsvColumnConfig{
			{Name: "a", Parser: "mock"},
			{Name: "b", Parser: "mock"},
		},
		ColumnIndexByName: map[string]int{
			"a": 0,
			"b": 1,
		},
	}

	csv, err := NewCsv(config, parsers)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("normal", func(t *testing.T) {
		// Testing a normal read
		expect := "test1"
		row, err := csv.Get(0)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		if result, _ := row.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}

		// Testing if the position in the file and buffer are properly set
		// when it has already been used once
		expect = "test3"
		row, err = csv.Get(12)
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		if result, _ := row.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
	})
	t.Run("parallel", func(t *testing.T) {
		// Executing two read operations in parallel to test the safety
		errorMessages1 := make(chan string)
		errorMessages2 := make(chan string)

		go (func() {
			expect := "test3"
			row, err := csv.Get(12)
			if err != nil {
				errorMessages1 <- fmt.Sprintf("Expected no error, got '%v'", err)
			}
			if result, _ := row.Get("a"); result != expect {
				errorMessages1 <- fmt.Sprintf("Expected '%v', got '%v'", expect, result)
			}
			close(errorMessages1)
		})()
		go (func() {
			expect := "test1"
			row, err := csv.Get(0)
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
		iterator, end, err := csv.IterateAll()
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		defer func() {
			if err := end(); err != nil {
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

		record, err := csv.Get(secondRow.Position())
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}

		expect := "test3"
		if result, _ := secondRow.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
		if result, _ := record.Get("a"); result != expect {
			t.Fatalf("Expected '%v', got '%v'", expect, result)
		}
	})
}

func TestCsvSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, err := createCsvTestFile(t, data)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		falseValue := false
		config := &CsvConfig{
			Path:             file.Name(),
			IgnoreFirstRow:   false,
			DieOnInputChange: &falseValue,
			Delimiter:        ",",
			Logger:           logrus.NewEntry(logrus.StandardLogger()),
			Columns: []*CsvColumnConfig{
				{Name: "a", Parser: "mock"},
			},
			ColumnIndexByName: map[string]int{
				"a": 0,
			},
		}

		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Fatal(err)
		}

		size, err := csv.Size()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Fatalf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestCsvIterateAll(t *testing.T) {
	testCases := []struct {
		name              string
		file              string
		expectedRows      [][]interface{}
		expectedPositions []int64
	}{
		{
			name:              "normal",
			file:              "test1,\"test2\"\n\"test\"\"3\",test 4",
			expectedRows:      [][]interface{}{{"test1", "test2"}, {"test\"3", "test 4"}},
			expectedPositions: []int64{0, 14},
		}, {
			name:              "empty",
			file:              "",
			expectedRows:      [][]interface{}{},
			expectedPositions: []int64{},
		}, {
			name:              "empty row",
			file:              "test1,test2\n\ntest3,test4",
			expectedRows:      [][]interface{}{{"test1", "test2"}, {"test3", "test4"}},
			expectedPositions: []int64{0, 12},
		}, {
			name:              "end after one row",
			file:              "test1,test2",
			expectedRows:      [][]interface{}{{"test1", "test2"}},
			expectedPositions: []int64{0},
		}, {
			name:              "too many columns",
			file:              "test1,test2\ntest3,test4,test5",
			expectedRows:      [][]interface{}{{"test1", "test2"}, {"test3", "test4", "test5"}},
			expectedPositions: []int64{0, 12},
		}, {
			name:              "not enough columns",
			file:              "test1,test2\ntest3",
			expectedRows:      [][]interface{}{{"test1", "test2"}, {"test3", nil}},
			expectedPositions: []int64{0, 12},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := createCsvTestFile(t, testCase.file)
			if err != nil {
				t.Fatalf("Unexpected error: '%+v'", err)
			}
			defer file.Close()

			parsers := parser.List{"mock": parser.NewMock()}

			logger := logrus.StandardLogger()
			logger.Level = logrus.ErrorLevel

			falseValue := false
			config := &CsvConfig{
				Path:             file.Name(),
				IgnoreFirstRow:   false,
				DieOnInputChange: &falseValue,
				Delimiter:        ",",
				Logger:           logrus.NewEntry(logger),
				Columns: []*CsvColumnConfig{
					{Name: "a", Parser: "mock"},
					{Name: "b", Parser: "mock"},
				},
				ColumnIndexByName: map[string]int{
					"a": 0,
					"b": 1,
				},
			}
			csv, err := NewCsv(config, parsers)
			if err != nil {
				t.Fatal(err)
			}

			iterator, end, err := csv.IterateAll()
			if err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
			defer func() {
				if err := end(); err != nil {
					t.Fatalf("Expected no error, got '%v'", err)
				}
			}()

			for i := 0; i < len(testCase.expectedRows); i++ {
				if record, err := iterator(); err != nil {
					t.Fatal(err)
				} else {
					if record == nil {
						break
					}

					for j, k := range []string{"a", "b"} {
						result, err := record.Get(k)
						if err != nil {
							t.Fatalf("Got error '%v', expected '%v' for cell [%v][%v]", err, testCase.expectedRows[i][j], i, j)
						}

						expected := testCase.expectedRows[i][j]
						if !(result == nil && expected == nil) && result != expected {
							t.Fatalf("Received '%v', expected '%v' for cell [%v][%v]", result, testCase.expectedRows[i][j], i, j)
						}
					}

					if got, expect := record.Position(), testCase.expectedPositions[i]; got != expect {
						t.Fatalf("Got position '%v', expected '%v' for row [%v]", got, expect, i)
					}
				}

				// Asserts that IterateAll does not fail with concurrent accesses
				if _, err := csv.reader.Seek(0, io.SeekStart); err != nil {
					t.Fatalf("Got error '%v'", err)
				}
			}
		})
	}
}

func TestCsvAutodetectColumns(t *testing.T) {
	parsers := parser.List{"string": parser.NewMock()}
	falseValue := false
	config := &CsvConfig{
		IgnoreFirstRow:    true,
		DieOnInputChange:  &falseValue,
		AutodetectColumns: true,
		Delimiter:         ",",
		Logger:            logrus.NewEntry(logrus.StandardLogger()),
	}

	t.Run("normal", func(t *testing.T) {
		file, err := createCsvTestFile(t, "test1,test2\n\ntest3,test4")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Fatal(err)
		}

		for index, name := range map[int]string{
			0: "test1",
			1: "test2",
		} {
			if index >= len(csv.config.Columns) {
				t.Fatalf("Expected to have a column indexed at '%v', got nothing", index)
			}

			column := csv.config.Columns[index]
			if column.Name != name {
				t.Fatalf("Expected to have a column named '%v' indexed at '%v', got '%v'", name, index, column.Name)
			}
			if column.Parser != "string" {
				t.Fatalf("Expected the column indexed at '%v' to have parser '%v', got '%v'", index, "string", column.Parser)
			}

			columnIndex, columnIndexExists := csv.config.ColumnIndexByName[name]
			if !columnIndexExists {
				t.Fatalf("Expected to have a column indexed under the name '%v', got nothing", name)
			}
			if columnIndex != index {
				t.Fatalf("Expected to have index '%v' for column '%v', got '%v'", index, name, columnIndex)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		file, err := createCsvTestFile(t, "test1,,test2\n\ntest3,test4")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		_, err = NewCsv(config, parsers)
		if err == nil {
			t.Fatalf("Expected to get an error, got '%v'", err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		file, err := createCsvTestFile(t, "test1,test1\n\ntest3,test4")
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		_, err = NewCsv(config, parsers)
		if err == nil {
			t.Fatalf("Expected to get an error, got '%v'", err)
		}
	})
}

func TestCsvOpen(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, err := createCsvTestFile(t, data)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		falseValue := false
		config := &CsvConfig{
			Path:             file.Name(),
			IgnoreFirstRow:   false,
			DieOnInputChange: &falseValue,
			Delimiter:        ",",
			Logger:           logrus.NewEntry(logrus.StandardLogger()),
			Columns: []*CsvColumnConfig{
				{Name: "a", Parser: "mock"},
			},
			ColumnIndexByName: map[string]int{
				"a": 0,
			},
		}

		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Fatal(err)
		}

		reader, _, _, file, err := csv.open()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if string(content) != data {
			t.Fatalf("Expected to receive '%v', got '%+v'", data, string(content))
		}

		if err := file.Close(); err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
	})
}
