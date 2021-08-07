package e2e

import (
	"testing"
)

func TestParameters(t *testing.T) {
	t.Run("word found with exact search", func(t *testing.T) {
		items := getListResponse(t, ServerUrl + "/?word=食べる")
		if got, expect := len(items), 1; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		itemMap, isMap := items[0].(map[string]interface{})
		if !isMap {
			t.Fatalf("The first item of the result is not an object: %v", items[0])
		}

		writing, writingIsString := itemMap["writing"].(string)
		if !writingIsString {
			t.Fatalf("The writing of the first item of the result is not a string: %v", itemMap["writing"])
		}
		if expect, got := "食べる", writing; expect != writing {
			t.Fatalf("Expected writing to be %v, got %v.", expect, got)
		}
	})
	t.Run("word not found with suffix", func(t *testing.T) {
		items := getListResponse(t, ServerUrl + "/?word=べる")
		if got, expect := len(items), 0; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("word not found with prefix", func(t *testing.T) {
		items := getListResponse(t, ServerUrl + "/?word=食べ")
		if got, expect := len(items), 0; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
}

// TODO test translation parameter
// TODO test query parameter
// TODO test multiple parameters
