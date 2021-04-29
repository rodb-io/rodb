package input

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"rodb.io/pkg/config"
	"rodb.io/pkg/record"
	"sync"
	"testing"
)

func createJsonTestFile(t *testing.T, data string) (*os.File, error) {
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

func createJsonTestInput(t *testing.T, data string) (*os.File, *Json, error) {
	file, err := createJsonTestFile(t, data)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	falseValue := false
	jsonConfig := &config.JsonInput{
		Path:             file.Name(),
		DieOnInputChange: &falseValue,
		Logger:           logrus.NewEntry(logrus.StandardLogger()),
	}

	json, err := NewJson(jsonConfig)
	if err != nil {
		return nil, nil, err
	}

	return file, json, nil
}

func TestJsonGet(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		file, json, err := createJsonTestInput(t, `
			{"a": "a0", "b": "b0"}
			{"a": "a1", "b": "b1"}
			{"a": "a2", "b": "b2"}
		`)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		// Testing a normal read
		row, err := json.Get(0)
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
		row, err = json.Get(26)
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
		file, json, err := createJsonTestInput(t, `
			{"a": "a0", "b": "b0"}
			{"a": "a1", "b": "b1"}
			{"a": "a2", "b": "b2"}
		`)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		// Executing two read operations in parallel to test the safety
		wait := sync.WaitGroup{}
		wait.Add(2)
		go (func() {
			expect := "a1"
			row, err := json.Get(26)
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
			row, err := json.Get(0)
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
		file, json, err := createJsonTestInput(t, `
			{"a": "a0", "b": "b0"}{"a": "a1", "b": "b1"}{"a": "a2", "b": "b2"}
		`)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		iterator, end, err := json.IterateAll()
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		defer func() {
			err := end()
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
		}()

		index := 0
		var secondRow record.Record = nil
		for {
			record, err := iterator()
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
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
			t.Errorf("Expected a record, got '%v'", secondRow)
		}

		record, err := json.Get(secondRow.Position())
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

func TestJsonSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, json, err := createJsonTestInput(t, data)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		size, err := json.Size()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Errorf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestJsonIterateAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		file, json, err := createJsonTestInput(t, `{"val": 1}{"val": 42}{"val": 123}`)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		iterator, end, err := json.IterateAll()
		if err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		defer func() {
			err := end()
			if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
		}()

		expectedVals := []float64{1, 42, 123}
		expectedPositions := []int64{0, 10, 21}
		for i := 0; i < len(expectedVals); i++ {
			if record, err := iterator(); err != nil {
				t.Error(err)
			} else {
				if record == nil {
					break
				}

				result, err := record.Get("val")
				if err != nil {
					t.Errorf("Got error '%v', expected '%v' for record %v", err, expectedVals[i], i)
					continue
				}

				expected := expectedVals[i]
				if result != expected {
					t.Errorf("Received '%v', expected '%v' for record %v", result, expectedVals[i], i)
				}

				if got, expect := record.Position(), expectedPositions[i]; got != expect {
					t.Errorf("Got position '%v', expected '%v' for record %v", got, expect, i)
				}
			}

			// Asserts that IterateAll does not fail with concurrent accesses
			json.reader.Seek(0, io.SeekStart)
		}
	})
}