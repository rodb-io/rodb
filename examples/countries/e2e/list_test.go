package e2e

import (
	"testing"
)

func TestList(t *testing.T) {
	waitForServer(t)
	t.Run("list", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/", &items)
		if got, expect := len(items), 262; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			if _, countryCodeIsString := itemMap["countryCode"].(string); !countryCodeIsString {
				t.Fatalf("countryCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["countryCode"])
			}
			if _, countryNameIsString := itemMap["countryName"].(string); !countryNameIsString {
				t.Fatalf("countryName property of the index %v of the result is not a string: %v", itemIndex, itemMap["countryName"])
			}
			if _, continentCodeIsString := itemMap["continentCode"].(string); !continentCodeIsString {
				t.Fatalf("continentCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["continentCode"])
			}
			if _, continentNameIsString := itemMap["continentName"].(string); !continentNameIsString {
				t.Fatalf("continentName property of the index %v of the result is not a string: %v", itemIndex, itemMap["continentName"])
			}
		}
	})
	t.Run("filter continentCode", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/?continentCode=EU", &items)
		if got, expect := len(items), 57; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			if _, countryCodeIsString := itemMap["countryCode"].(string); !countryCodeIsString {
				t.Fatalf("countryCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["countryCode"])
			}
			if _, countryNameIsString := itemMap["countryName"].(string); !countryNameIsString {
				t.Fatalf("countryName property of the index %v of the result is not a string: %v", itemIndex, itemMap["countryName"])
			}
			if _, continentCodeIsString := itemMap["continentCode"].(string); !continentCodeIsString {
				t.Fatalf("continentCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["continentCode"])
			}
			if _, continentNameIsString := itemMap["continentName"].(string); !continentNameIsString {
				t.Fatalf("continentName property of the index %v of the result is not a string: %v", itemIndex, itemMap["continentName"])
			}

			if itemMap["continentCode"] != "EU" {
				t.Fatalf("expected continentCode of the index %v of the result to be 'EU', got '%v'", itemIndex, itemMap["continentCode"])
			}
		}
	})
	t.Run("filter countryCode", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/?countryCode=FR", &items)
		if got, expect := len(items), 1; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			if _, countryCodeIsString := itemMap["countryCode"].(string); !countryCodeIsString {
				t.Fatalf("countryCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["countryCode"])
			}
			if _, countryNameIsString := itemMap["countryName"].(string); !countryNameIsString {
				t.Fatalf("countryName property of the index %v of the result is not a string: %v", itemIndex, itemMap["countryName"])
			}
			if _, continentCodeIsString := itemMap["continentCode"].(string); !continentCodeIsString {
				t.Fatalf("continentCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["continentCode"])
			}
			if _, continentNameIsString := itemMap["continentName"].(string); !continentNameIsString {
				t.Fatalf("continentName property of the index %v of the result is not a string: %v", itemIndex, itemMap["continentName"])
			}

			if itemMap["countryCode"] != "FR" {
				t.Fatalf("expected countryCode of the index %v of the result to be 'FR', got '%v'", itemIndex, itemMap["countryCode"])
			}
		}
	})
}
