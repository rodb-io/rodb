package index

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"rods/pkg/input"
	"testing"
)

func TestNewFromConfigs(t *testing.T) {
	t.Run("has default noop index", func(t *testing.T) {
		indexes, err := NewFromConfigs(
			map[string]config.Index{},
			input.List{},
			logrus.StandardLogger(),
		)
		if err != nil {
			t.Errorf("Expected no error, got '%+v'", err)
		}

		if _, ok := indexes[""]; !ok {
			t.Errorf("Expected to have a default index at empty the string key, got nothing")
		}
	})
}
