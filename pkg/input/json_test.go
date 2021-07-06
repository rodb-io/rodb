package input

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"rodb.io/pkg/config"
	"rodb.io/pkg/record"
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
	jsonConfig := &JsonConfig{
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
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		// Testing a normal read
		row, err := json.Get(0)
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
		row, err = json.Get(26)
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
		file, json, err := createJsonTestInput(t, `
			{"a": "a0", "b": "b0"}
			{"a": "a1", "b": "b1"}
			{"a": "a2", "b": "b2"}
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		// Executing two read operations in parallel to test the safety
		errorMessages1 := make(chan string)
		errorMessages2 := make(chan string)

		go (func() {
			expect := "a1"
			row, err := json.Get(26)
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
			row, err := json.Get(0)
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
		file, json, err := createJsonTestInput(t, `
			{"a": "a0", "b": "b0"}{"a": "a1", "b": "b1"}{"a": "a2", "b": "b2"}
		`)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		iterator, end, err := json.IterateAll()
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

		record, err := json.Get(secondRow.Position())
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

func TestJsonSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := "Hello World!"
		file, json, err := createJsonTestInput(t, data)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		size, err := json.Size()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Fatalf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestJsonIterateAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		file, json, err := createJsonTestInput(t, `{"val": 1}{"val": 42}{"val": 123}`)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}
		defer file.Close()

		iterator, end, err := json.IterateAll()
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		defer func() {
			if err := end(); err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
		}()

		expectedVals := []float64{1, 42, 123}
		expectedPositions := []int64{0, 10, 21}
		for i := 0; i < len(expectedVals); i++ {
			if record, err := iterator(); err != nil {
				t.Fatal(err)
			} else {
				if record == nil {
					break
				}

				result, err := record.Get("val")
				if err != nil {
					t.Fatalf("Got error '%v', expected '%v' for record %v", err, expectedVals[i], i)
				}

				expected := expectedVals[i]
				if result != expected {
					t.Fatalf("Received '%v', expected '%v' for record %v", result, expectedVals[i], i)
				}

				if got, expect := record.Position(), expectedPositions[i]; got != expect {
					t.Fatalf("Got position '%v', expected '%v' for record %v", got, expect, i)
				}
			}

			// Asserts that IterateAll does not fail with concurrent accesses
			if _, err := json.reader.Seek(0, io.SeekStart); err != nil {
				t.Fatalf("Got error '%v'", err)
			}
		}
	})
}
