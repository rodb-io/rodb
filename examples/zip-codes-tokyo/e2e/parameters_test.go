package e2e

import (
	"testing"
)

func TestParameters(t *testing.T) {
	waitForServer(t)
	t.Run("municipality found with exact search", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?municipality=世田谷区", &items)

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			if got, expect := itemMap["municipality"], "世田谷区"; got != expect {
				t.Fatalf("municipality property of the index %v of the result does not match the search parameter: %#v", itemIndex, itemMap)
			}
		}
	})
	t.Run("municipality not found with suffix", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?municipality=区", &items)
		if got, expect := len(items), 0; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("municipality not found with prefix", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?municipality=世田", &items)
		if got, expect := len(items), 0; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("get item", func(t *testing.T) {
		item := map[string]interface{}{}
		getResponse(t, ServerUrl+"/zip-codes/1006822", &item)

		if got, expect := item["zipCode"], "1006822"; got != expect {
			t.Fatalf("zipCode property of the result does not match the input parameter: %#v", item)
		}
		if got, expect := item["municipality"], "千代田区"; got != expect {
			t.Fatalf("municipality property of the result does not match the input parameter: %#v", item)
		}
	})
}
