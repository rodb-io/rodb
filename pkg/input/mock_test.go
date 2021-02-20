package input

import (
	"errors"
	"rods/pkg/record"
	"testing"
)

func TestMockGet(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedRecord := record.NewStringColumnsMock(map[string]string{
			"col": "value",
		}, 0)
		mock := NewMock([]IterateAllResult{{Record: expectedRecord}})

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
		mock := NewMock([]IterateAllResult{{Error: expectedError}})

		_, err := mock.Get(0)
		if err != expectedError {
			t.Errorf("Expected error '%+v', got '%+v'", expectedError, err)
		}
	})
	t.Run("unexpected error", func(t *testing.T) {
		mock := NewMock([]IterateAllResult{})

		_, err := mock.Get(0)
		if err == nil {
			t.Errorf("Expected an error, got '%+v'", err)
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
		mock := NewMock(data)

		channel := mock.IterateAll()

		for i := 0; i < len(data); i++ {
			if result := <-channel; result != data[i] {
				t.Errorf("Expected %+v, got %+v", data[i], result)
			}
		}
	})
}
