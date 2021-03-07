package output

import (
	"testing"
)

func TestMock(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mock := NewMock()
		_ = mock
	})
}

func TestMockClose(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mock := NewMock()

		err := mock.Close()
		if err != nil {
			t.Errorf("Unexpected error: '%+v'", err)
		}
	})
}
