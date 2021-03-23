package config

import (
	"testing"
)

func TestConfigCheckDuplicateEndpointsPerService(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		config := &Config{
			Outputs: map[string]Output{
				"Test": {
					JsonArray: &JsonArrayOutput{
						Services: []string{"test"},
						Endpoint: "/",
					},
				},
			},
		}
		if err := config.checkDuplicateEndpointsPerService(); err != nil {
			t.Errorf("Expected to not get an error, got %v", err)
		}
	})
	t.Run("duplicates", func(t *testing.T) {
		config := &Config{
			Outputs: map[string]Output{
				"Test": {
					JsonArray: &JsonArrayOutput{
						Services: []string{"test"},
						Endpoint: "/",
					},
				},
				"Test2": {
					JsonArray: &JsonArrayOutput{
						Services: []string{"test"},
						Endpoint: "/",
					},
				},
			},
		}
		if err := config.checkDuplicateEndpointsPerService(); err == nil {
			t.Errorf("Expected to get an error, got nil")
		}
	})
	t.Run("empty", func(t *testing.T) {
		config := &Config{
			Outputs: map[string]Output{},
		}
		if err := config.checkDuplicateEndpointsPerService(); err != nil {
			t.Errorf("Expected to not get an error, got %v", err)
		}
	})
}
