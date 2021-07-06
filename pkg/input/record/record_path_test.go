package record

import (
	"fmt"
	"testing"
)

func TestGetSubValue(t *testing.T) {
	testCases := []struct {
		name        string
		data        interface{}
		path        []string
		expectError bool
		expectValue interface{}
	}{
		{
			name:        "primitive",
			data:        "test",
			path:        []string{},
			expectError: false,
			expectValue: "test",
		}, {
			name:        "primitive in array",
			data:        []interface{}{"test"},
			path:        []string{"0"},
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in object",
			data: map[string]interface{}{
				"key": "test",
			},
			path:        []string{"key"},
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in array of objects",
			data: []interface{}{
				map[string]interface{}{
					"key": "test",
				},
			},
			path:        []string{"0", "key"},
			expectError: false,
			expectValue: "test",
		}, {
			name: "primitive in array in an object",
			data: map[string]interface{}{
				"key": []interface{}{"test"},
			},
			path:        []string{"key", "0"},
			expectError: false,
			expectValue: "test",
		}, {
			name:        "array",
			data:        []interface{}{"test"},
			path:        []string{},
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name: "array in array",
			data: []interface{}{
				42,
				[]interface{}{"test"},
			},
			path:        []string{"1"},
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name: "array in object",
			data: map[string]interface{}{
				"key": []interface{}{"test"},
			},
			path:        []string{"key"},
			expectError: false,
			expectValue: []string{"test"},
		}, {
			name: "object",
			data: map[string]interface{}{
				"key": "test",
			},
			path:        []string{},
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name: "object in array",
			data: []interface{}{
				map[string]interface{}{
					"key": "test",
				},
			},
			path:        []string{"0"},
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name: "object in object",
			data: map[string]interface{}{
				"keyRoot": map[string]interface{}{
					"key": "test",
				},
			},
			path:        []string{"keyRoot"},
			expectError: false,
			expectValue: map[string]interface{}{"key": "test"},
		}, {
			name:        "index out of range",
			data:        []interface{}{"a", "b"},
			path:        []string{"5"},
			expectError: false,
			expectValue: nil,
		}, {
			name: "missing property",
			data: map[string]interface{}{
				"a": nil,
				"b": 42,
			},
			path:        []string{"nope"},
			expectError: false,
			expectValue: nil,
		}, {
			name:        "wrong path from root",
			data:        "test",
			path:        []string{"a", "b"},
			expectError: true,
		}, {
			name:        "non-numeric key",
			data:        []interface{}{"test"},
			path:        []string{"key"},
			expectError: true,
		}, {
			name: "wrong path",
			data: map[string]interface{}{
				"key": "value",
			},
			path:        []string{"key", "test", "42"},
			expectError: true,
		}, {
			name: "deep path",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": map[string]interface{}{
								"e": map[string]interface{}{
									"f": map[string]interface{}{
										"g": true,
									},
								},
							},
						},
					},
				},
			},
			path:        []string{"a", "b", "c", "d", "e", "f", "g"},
			expectError: false,
			expectValue: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := GetSubValue(testCase.data, testCase.path)
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
