package input

import (
	"errors"
	"rods/pkg/record"
	"testing"
)

func TestMockIterateAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := []IterateAllResult{
			{Record: record.NewSingleStringColumnMock("col", "value", 0)},
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
