package input

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"rods/pkg/config"
	"rods/pkg/parser"
	"sync"
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

func TestCsvHasColumn(t *testing.T) {
	file, err := createCsvTestFile(t, "")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	defer file.Close()

	parsers := parser.List{"mock": parser.NewMock()}

	config := &config.CsvInput{
		Path:           file.Name(),
		IgnoreFirstRow: false,
		Delimiter:      ",",
		Logger:         logrus.NewEntry(logrus.StandardLogger()),
		Columns: []*config.CsvInputColumn{
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
		t.Error(err)
	}

	t.Run("true", func(t *testing.T) {
		if !csv.HasColumn("a") {
			t.Errorf("Expected to have column 'a', got false")
		}
		if !csv.HasColumn("b") {
			t.Errorf("Expected to have column 'b', got false")
		}
	})
	t.Run("false", func(t *testing.T) {
		if csv.HasColumn("wrong") {
			t.Errorf("Expected to not have column 'wrong', got true")
		}
	})
}

func TestCsvGet(t *testing.T) {
	file, err := createCsvTestFile(t, "test1,test2\n\ntest3")
	if err != nil {
		t.Errorf("Unexpected error: '%+v'", err)
	}
	defer file.Close()

	parsers := parser.List{"mock": parser.NewMock()}

	config := &config.CsvInput{
		Path:           file.Name(),
		IgnoreFirstRow: false,
		Delimiter:      ",",
		Logger:         logrus.NewEntry(logrus.StandardLogger()),
		Columns: []*config.CsvInputColumn{
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
		t.Error(err)
	}

	t.Run("normal", func(t *testing.T) {
		// Testing a normal read
		expect := "test1"
		row, err := csv.Get(0)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if result, _ := row.Get("a"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}

		// Testing if the position in the file and buffer are properly set
		// when it has already been used once
		expect = "test3"
		row, err = csv.Get(12)
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		if result, _ := row.Get("a"); result != expect {
			t.Errorf("Expected '%v', got '%v'", expect, result)
		}
	})
	t.Run("parallel", func(t *testing.T) {
		// Executing two read operations in parallel to test the safety
		wait := sync.WaitGroup{}
		wait.Add(2)
		go (func() {
			expect := "test3"
			row, err := csv.Get(12)
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
			if result, _ := row.Get("a"); result != expect {
				t.Errorf("Expected '%v', got '%v'", expect, result)
			}
			wait.Done()
		})()
		go (func() {
			expect := "test1"
			row, err := csv.Get(0)
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
}

func TestCsvSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, err := createCsvTestFile(t, data)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		config := &config.CsvInput{
			Path:           file.Name(),
			IgnoreFirstRow: false,
			Delimiter:      ",",
			Logger:         logrus.NewEntry(logrus.StandardLogger()),
			Columns: []*config.CsvInputColumn{
				{Name: "a", Parser: "mock"},
			},
			ColumnIndexByName: map[string]int{
				"a": 0,
			},
		}

		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Error(err)
		}

		size, err := csv.Size()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Errorf("Expected to get a size of '%v', got '%+v'", len(data), size)
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
				t.Errorf("Unexpected error: '%+v'", err)
			}
			defer file.Close()

			parsers := parser.List{"mock": parser.NewMock()}

			config := &config.CsvInput{
				Path:           file.Name(),
				IgnoreFirstRow: false,
				Delimiter:      ",",
				Logger:         logrus.NewEntry(logrus.StandardLogger()),
				Columns: []*config.CsvInputColumn{
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
				t.Error(err)
			}

			resultsChannel := csv.IterateAll()
			for i := 0; i < len(testCase.expectedRows); i++ {
				if result := <-resultsChannel; result.Error != nil {
					t.Error(err)
				} else {
					for j, k := range []string{"a", "b"} {
						result, err := result.Record.Get(k)
						if err != nil {
							t.Errorf("Got error '%v', expected '%v' for cell [%v][%v]", err, testCase.expectedRows[i][j], i, j)
							continue
						}

						expected := testCase.expectedRows[i][j]
						if !(result == nil && expected == nil) && result != expected {
							t.Errorf("Received '%v', expected '%v' for cell [%v][%v]", result, testCase.expectedRows[i][j], i, j)
						}
					}

					if got, expect := result.Record.Position(), testCase.expectedPositions[i]; got != expect {
						t.Errorf("Got position '%v', expected '%v' for row [%v]", got, expect, i)
					}
				}

				// Asserts that IterateAll does not fail with concurrent accesses
				csv.reader.Seek(0, io.SeekStart)
			}
		})
	}
}

func TestCsvAutodetectColumns(t *testing.T) {
	parsers := parser.List{"string": parser.NewMock()}
	config := &config.CsvInput{
		IgnoreFirstRow:    true,
		AutodetectColumns: true,
		Delimiter:         ",",
		Logger:            logrus.NewEntry(logrus.StandardLogger()),
	}

	t.Run("normal", func(t *testing.T) {
		file, err := createCsvTestFile(t, "test1,test2\n\ntest3,test4")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Error(err)
		}

		for index, name := range map[int]string{
			0: "test1",
			1: "test2",
		} {
			if index >= len(csv.config.Columns) {
				t.Errorf("Expected to have a column indexed at '%v', got nothing", index)
			}

			column := csv.config.Columns[index]
			if column.Name != name {
				t.Errorf("Expected to have a column named '%v' indexed at '%v', got '%v'", name, index, column.Name)
			}
			if column.Parser != "string" {
				t.Errorf("Expected the column indexed at '%v' to have parser '%v', got '%v'", index, "string", column.Parser)
			}

			columnIndex, columnIndexExists := csv.config.ColumnIndexByName[name]
			if !columnIndexExists {
				t.Errorf("Expected to have a column indexed under the name '%v', got nothing", name)
			}
			if columnIndex != index {
				t.Errorf("Expected to have index '%v' for column '%v', got '%v'", index, name, columnIndex)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		file, err := createCsvTestFile(t, "test1,,test2\n\ntest3,test4")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		_, err = NewCsv(config, parsers)
		if err == nil {
			t.Errorf("Expected to get an error, got '%v'", err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		file, err := createCsvTestFile(t, "test1,test1\n\ntest3,test4")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		config.Path = file.Name()
		_, err = NewCsv(config, parsers)
		if err == nil {
			t.Errorf("Expected to get an error, got '%v'", err)
		}
	})
}

func TestCsvOpen(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, err := createCsvTestFile(t, data)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		config := &config.CsvInput{
			Path:           file.Name(),
			IgnoreFirstRow: false,
			Delimiter:      ",",
			Logger:         logrus.NewEntry(logrus.StandardLogger()),
			Columns: []*config.CsvInputColumn{
				{Name: "a", Parser: "mock"},
			},
			ColumnIndexByName: map[string]int{
				"a": 0,
			},
		}

		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Error(err)
		}

		reader, _, file, err := csv.open()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if string(content) != data {
			t.Errorf("Expected to receive '%v', got '%+v'", data, string(content))
		}

		err = file.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
	})
}

func TestCsvWatch(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		file, err := createCsvTestFile(t, "initial content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		parsers := parser.List{"mock": parser.NewMock()}

		trueValue := true
		config := &config.CsvInput{
			Path:             file.Name(),
			IgnoreFirstRow:   false,
			DieOnInputChange: &trueValue,
			Delimiter:        ",",
			Logger:           logrus.NewEntry(logrus.StandardLogger()),
			Columns: []*config.CsvInputColumn{
				{Name: "a", Parser: "mock"},
			},
			ColumnIndexByName: map[string]int{
				"a": 0,
			},
		}

		csv, err := NewCsv(config, parsers)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieWaiter := &sync.WaitGroup{}
		dieCount := 0
		csv.config.Logger.Logger.ExitFunc = func(exitCode int) {
			dieCount++
			dieWaiter.Done()
		}

		dieWaiter.Add(1)
		_, err = file.WriteString("changed content")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieWaiter.Wait()
		if dieCount <= 0 {
			t.Errorf("Expected the process to exit, got '%v' calls to Exit", dieCount)
		}

		err = csv.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		dieCount = 0
		dieWaiter.Add(1)
		_, err = file.WriteString("changed content again")
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if dieCount != 0 {
			t.Errorf("Expected the process not to exit, got '%v' calls to Exit", dieCount)
		}
	})
}
