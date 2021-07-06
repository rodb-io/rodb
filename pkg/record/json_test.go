package record

import (
	"fmt"
	"testing"
)

func TestJsonAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		data := map[string]interface{}{
			"test": 123,
		}
		record := NewJson(&JsonConfig{}, data, 0)
		got, err := record.All()
		if err != nil {
			t.Fatalf("Got error: '%v'", err)
		}

		if expect := 123; got["test"] != expect {
			t.Fatalf("Expected to get '%v', got '%v'", expect, data["col_b"])
		}
	})
}

func TestJsonGet(t *testing.T) {
	testCases := []struct {
		name        string
		data        map[string]interface{}
		path        string
		expectError bool
		expectValue interface{}
	}{
		{
			name:        "primitive",
			data:        map[string]interface{}{"col": "test"},
			path:        "col",
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in array",
			data: map[string]interface{}{
				"col": []interface{}{
					"test",
				},
			},
			path:        "col.0",
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in object",
			data: map[string]interface{}{
				"col": map[string]interface{}{
					"key": "test",
				},
			},
			path:        "col.key",
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in array of objects",
			data: map[string]interface{}{
				"col": []interface{}{
					map[string]interface{}{
						"key": "test",
					},
				},
			},
			path:        "col.0.key",
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in array in an object",
			data: map[string]interface{}{
				"col": map[string]interface{}{
					"key": []interface{}{
						"test",
					},
				},
			},
			path:        "col.key.0",
			expectError: false,
			expectValue: "test",
		}, {
			name: "array",
			data: map[string]interface{}{
				"col": []interface{}{"test"},
			},
			path:        "col",
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name: "array in array",
			data: map[string]interface{}{
				"col": []interface{}{
					nil,
					[]string{"test"},
				},
			},
			path:        "col.1",
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name: "object",
			data: map[string]interface{}{
				"col": map[string]interface{}{
					"key": "test",
				},
			},
			path:        "col",
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name: "index out of range",
			data: map[string]interface{}{
				"col": []interface{}{
					42,
				},
			},
			path:        "col.5",
			expectError: false,
			expectValue: nil,
		}, {
			name: "missing property",
			data: map[string]interface{}{
				"col": map[string]interface{}{},
			},
			path:        "col.nope",
			expectError: false,
			expectValue: nil,
		}, {
			name: "non-numeric key",
			data: map[string]interface{}{
				"col": []interface{}{},
			},
			path:        "col.key",
			expectError: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			record := NewJson(&JsonConfig{}, testCase.data, 0)

			got, err := record.Get(testCase.path)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("Expected error, got: '%v'", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: '%v'", err)
				}
			}

			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", testCase.expectValue) {
				t.Fatalf("Expected to get '%v', got '%v'", testCase.expectValue, got)
			}
		})
	}
}
