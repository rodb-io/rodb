package input

import (
	"errors"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"testing"
)

func TestMockHasColumn(t *testing.T) {
	columnName := "col"
	expectedRecord := record.NewStringColumnsMock(map[string]string{
		columnName: "value",
	}, 0)
	mock := NewMock(parser.NewMock(), []IterateAllResult{{Record: expectedRecord}})

	t.Run("true", func(t *testing.T) {
		if !mock.HasColumn(columnName) {
			t.Errorf("Expected to have column '%v', got false", columnName)
		}
	})
	t.Run("false", func(t *testing.T) {
		if mock.HasColumn("wrong_" + columnName) {
			t.Errorf("Expected to not have column 'wrong_%v', got true", columnName)
		}
	})
}

func TestMockGet(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedRecord := record.NewStringColumnsMock(map[string]string{
			"col": "value",
		}, 0)
		mock := NewMock(parser.NewMock(), []IterateAllResult{{Record: expectedRecord}})

		record, err := mock.Get(0)
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if record != expectedRecord {
			t.Errorf("Expected %+v, got %+v", expectedRecord, record)
		}
	})
	t.Run("expected error", func(t *testing.T) {
		expectedError := errors.New("Test error")
		mock := NewMock(parser.NewMock(), []IterateAllResult{{Error: expectedError}})

		_, err := mock.Get(0)
		if err != expectedError {
			t.Errorf("Expected error '%+v', got '%+v'", expectedError, err)
		}
	})
	t.Run("unexpected error", func(t *testing.T) {
		mock := NewMock(parser.NewMock(), []IterateAllResult{})

		_, err := mock.Get(0)
		if err == nil {
			t.Errorf("Expected an error, got '%+v'", err)
		}
	})
}

func TestMockSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedRecord := record.NewStringColumnsMock(map[string]string{
			"col": "value",
		}, 0)
		data := []IterateAllResult{
			{Record: expectedRecord},
		}
		mock := NewMock(parser.NewMock(), data)

		size, err := mock.Size()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Errorf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestMockIterateAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := []IterateAllResult{
			{Record: record.NewStringColumnsMock(map[string]string{
				"col": "value",
			}, 0)},
			{Error: errors.New("Test error")},
		}
		mock := NewMock(parser.NewMock(), data)

		channel := mock.IterateAll()

		for i := 0; i < len(data); i++ {
			if result := <-channel; result != data[i] {
				t.Errorf("Expected %+v, got %+v", data[i], result)
			}
		}
	})
}
