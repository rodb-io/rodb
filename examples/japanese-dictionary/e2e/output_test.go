package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestOutput(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		items := getListResponse(t, ServerUrl + "/")
		if got, expect := len(items), 100; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		writingCount := 0
		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			writing, writingIsString := itemMap["reading"].(string)
			if !writingIsString {
				t.Fatalf("reading property of the index %v of the result is not a string: %v", itemIndex, itemMap["reading"])
			}
			if writing == "" {
				t.Fatalf("reading property of the index %v of the result is empty", itemIndex)
			}

			translation, writingIsString := itemMap["translation"].(string)
			if !writingIsString {
				t.Fatalf("translation property of the index %v of the result is not a string: %v", itemIndex, itemMap["translation"])
			}
			if translation == "" {
				t.Fatalf("translation property of the index %v of the result is empty", itemIndex)
			}

			reading, writingIsString := itemMap["writing"].(string)
			if !writingIsString {
				t.Fatalf("writing property of the index %v of the result is not a string: %v", itemIndex, itemMap["writing"])
			}
			if reading != "" {
				// The writing is optional, but rarely empty
				writingCount++
			}
		}

		if writingCount == 0 {
			t.Fatalf("The list only contains empty writings")
		}
	})
	t.Run("custom limit", func(t *testing.T) {
		itemsDefault := getListResponse(t, ServerUrl + "/")
		if got, expect := len(itemsDefault), 100; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		itemsCustom := getListResponse(t, ServerUrl + "/?limit=10")
		if got, expect := len(itemsCustom), 10; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		if fmt.Sprintf("%#v", itemsCustom[0]) != fmt.Sprintf("%#v", itemsDefault[0]) {
			t.Fatalf("Expected the results to be identical. Expected %#v, got %#v", itemsDefault[0], itemsCustom[0])
		}
	})
	t.Run("limit > max", func(t *testing.T) {
		items := getListResponse(t, ServerUrl + "/?limit=2000")
		if got, expect := len(items), 1000; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("limit < 0", func(t *testing.T) {
		getErrorResponse(t, ServerUrl + "/?limit=-123", http.StatusInternalServerError)
	})
	t.Run("limit = 0", func(t *testing.T) {
		getErrorResponse(t, ServerUrl + "/?limit=0", http.StatusInternalServerError)
	})
}

// TODO have a different return status for the validation errors than 500

// TODO test word parameter
// TODO test translation parameter
// TODO test query parameter
// TODO test multiple parameters
