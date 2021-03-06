package input

import (
	recordPackage "github.com/rodb-io/rodb/pkg/input/record"
	"github.com/rodb-io/rodb/pkg/parser"
	"testing"
)

func TestMockGet(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedRecord := recordPackage.NewStringPropertiesMockRecord(map[string]string{
			"col": "value",
		}, 0)
		mock := NewMock(parser.NewMock(), []recordPackage.Record{expectedRecord})

		record, err := mock.Get(0)
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if record != expectedRecord {
			t.Fatalf("Expected %+v, got %+v", expectedRecord, record)
		}
	})
}

func TestMockSize(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		expectedRecord := recordPackage.NewStringPropertiesMockRecord(map[string]string{
			"col": "value",
		}, 0)
		data := []recordPackage.Record{expectedRecord}
		mock := NewMock(parser.NewMock(), data)

		size, err := mock.Size()
		if err != nil {
			t.Fatalf("Unexpected error: '%+v'", err)
		}

		if size != int64(len(data)) {
			t.Fatalf("Expected to get a size of '%v', got '%+v'", len(data), size)
		}
	})
}

func TestMockIterateAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := []recordPackage.Record{
			recordPackage.NewStringPropertiesMockRecord(map[string]string{
				"col": "value",
			}, 0),
			recordPackage.NewStringPropertiesMockRecord(map[string]string{
				"col": "value",
			}, 1),
		}
		mock := NewMock(parser.NewMock(), data)

		iterator, end, err := mock.IterateAll()
		if err != nil {
			t.Fatalf("Expected no error, got '%v'", err)
		}
		defer func() {
			if err := end(); err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}
		}()

		for i := 0; i < len(data); i++ {
			record, err := iterator()
			if err != nil {
				t.Fatalf("Expected no error, got '%v'", err)
			}

			if record != data[i] {
				t.Fatalf("Expected %+v, got %+v", data[i], record)
			}
		}
	})
}
