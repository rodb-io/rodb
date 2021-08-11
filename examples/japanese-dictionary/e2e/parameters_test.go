package e2e

import (
	"net/url"
	"strings"
	"testing"
)

func TestParameters(t *testing.T) {
	waitForServer(t)
	t.Run("word found with exact search", func(t *testing.T) {
		items := getListResponse(t, ServerUrl+"/?word=食べる")
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
		items := getListResponse(t, ServerUrl+"/?word=べる")
		if got, expect := len(items), 0; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("word not found with prefix", func(t *testing.T) {
		items := getListResponse(t, ServerUrl+"/?word=食べ")
		if got, expect := len(items), 0; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("translation wildcard", func(t *testing.T) {
		items := getListResponse(t, ServerUrl+"/?translation=table")

		found := false
		expect := "wet towel (supplied at table)"
		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("The item at index %v is not an object: %v", itemIndex, item)
			}

			if itemMap["translation"] == expect {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Expected to find a result with the translation '%v', but it was not found among %v results.", found, len(items))
		}
	})
	t.Run("translation full string", func(t *testing.T) {
		items := getListResponse(t, ServerUrl+"/?translation="+url.QueryEscape("wet towel (supplied at table)"))
		if got, expect := len(items), 1; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("word and translation", func(t *testing.T) {
		items := getListResponse(t, ServerUrl+"/?word=食べる&translation="+url.QueryEscape("to eat"))
		if got, expect := len(items), 1; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("match", func(t *testing.T) {
		items := getListResponse(t, ServerUrl+"/?query="+url.QueryEscape("(translation: trip AND translation: day) OR translation:work hard"))

		foundDayTrip := false
		foundWorkHard := false
		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("The item at index %v is not an object: %v", itemIndex, item)
			}

			translation, translationIsString := itemMap["translation"].(string)
			if !translationIsString {
				t.Fatalf("The translation of the item at index %v of the result is not a string: %v", itemIndex, itemMap["translation"])
			}

			if strings.Contains(translation, "day trip") {
				foundDayTrip = true
			}
			if strings.Contains(translation, "work hard") {
				foundWorkHard = true
			}
			if foundDayTrip && foundWorkHard {
				break
			}
		}

		if !foundDayTrip {
			t.Fatalf("Expected to find at least one result containing 'day trip', but it was not found among %v results.", len(items))
		}
		if !foundWorkHard {
			t.Fatalf("Expected to find at least one result containing 'work hard', but it was not found among %v results.", len(items))
		}
	})
}
