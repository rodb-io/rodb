package parser

import (
	"github.com/sirupsen/logrus"
	"rods/pkg/config"
	"testing"
)

func TestNewFromConfigs(t *testing.T) {
	t.Run("has default dumb index", func(t *testing.T) {
		indexes, err := NewFromConfigs(
			map[string]config.Parser{},
			logrus.StandardLogger(),
		)
		if err != nil {
			t.Errorf("Expected no error, got '%+v'", err)
		}

		for _, key := range []string{"string", "integer", "float", "boolean"} {
			if _, ok := indexes[key]; !ok {
				t.Errorf("Expected to have a default index at key '%+v', got nothing", key)
			}
		}
	})
}
