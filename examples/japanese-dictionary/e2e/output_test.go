package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
)

func getListOutput(t *testing.T, path string) []interface{} {
	response, err := getClient().Get(ServerUrl + path)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got, expect := response.StatusCode, http.StatusOK; got != expect {
		t.Fatalf("Got return status %v, expected %v", got, expect)
	}

	body := []interface{}{}
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&body); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	return body
}

func TestOutput(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		items := getListOutput(t, "/")
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
}

// TODO test setting limit manually
// TODO test setting a limit over 1000
// TODO test setting a limit <= 0
// TODO test word parameter
// TODO test translation parameter
// TODO test query parameter
// TODO test multiple parameters
