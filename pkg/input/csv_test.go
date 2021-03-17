package input

import (
	"github.com/sirupsen/logrus"
	"io"
	"rods/pkg/config"
	"rods/pkg/parser"
	"rods/pkg/source"
	"sync"
	"testing"
)

func TestCsvGet(t *testing.T) {
	parsers := parser.List{"mock": parser.NewMock()}
	sources := source.List{"mock": source.NewMock("test1,test2\n\ntest3")}

	config := &config.CsvInput{
		Path:           "test",
		Source:         "mock",
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

	csv, err := NewCsv(config, sources, parsers)
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
		parsers := parser.List{"mock": parser.NewMock()}
		data := "Hello World !"
		mockSource := source.NewMock(data)
		sources := source.List{"mock": mockSource}

		config := &config.CsvInput{
			Path:           "test",
			Source:         "mock",
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

		csv, err := NewCsv(config, sources, parsers)
		if err != nil {
			t.Error(err)
		}

		size, err := csv.Size("test")
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
			parsers := parser.List{"mock": parser.NewMock()}
			sources := source.List{"mock": source.NewMock(testCase.file)}

			config := &config.CsvInput{
				Path:           "test",
				Source:         "mock",
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
			csv, err := NewCsv(config, sources, parsers)
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
				csv.sourceReader.Seek(0, io.SeekStart)
			}
		})
	}
}

func TestCsvWatch(t *testing.T) {
	parsers := parser.List{"mock": parser.NewMock()}
	mockSource := source.NewMock("")
	sources := source.List{"mock": mockSource}

	config := &config.CsvInput{
		Path:           "test",
		Source:         "mock",
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

	csv, err := NewCsv(config, sources, parsers)
	if err != nil {
		t.Error(err)
	}

	t.Run("normal", func(t *testing.T) {
		callCount := 0
		watcher := &source.Watcher{
			OnChange: func() {
				callCount++
			},
		}

		err := csv.Watch(watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if callCount != 0 {
			t.Errorf("Expected the function to not be called, got '%v'", callCount)
		}

		mockSource.TriggerWatchers()
		if callCount != 1 {
			t.Errorf("Expected the function to be called once, got '%v'", callCount)
		}

		callCount = 0
		err = csv.CloseWatcher(watcher)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		mockSource.TriggerWatchers()
		if callCount != 0 {
			t.Errorf("Expected the function to not be called, got '%v'", callCount)
		}
	})
}
