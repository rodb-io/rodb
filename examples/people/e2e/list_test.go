package e2e

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	waitForServer(t)
	t.Run("list", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/people", &items)
		if got, expect := len(items), 100; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			if _, emailIsString := itemMap["email"].(string); !emailIsString {
				t.Fatalf("email property of the index %v of the result is not a string: %v", itemIndex, itemMap["email"])
			}
			if _, firstNameIsString := itemMap["firstName"].(string); !firstNameIsString {
				t.Fatalf("firstName property of the index %v of the result is not a string: %v", itemIndex, itemMap["firstName"])
			}
			if _, genderIsString := itemMap["gender"].(string); !genderIsString {
				t.Fatalf("gender property of the index %v of the result is not a string: %v", itemIndex, itemMap["gender"])
			}
			if _, idIsFloat := itemMap["id"].(float64); !idIsFloat {
				t.Fatalf("id property of the index %v of the result is not a float64: %v", itemIndex, itemMap["id"])
			}
			if _, lastNameIsString := itemMap["lastName"].(string); !lastNameIsString {
				t.Fatalf("lastName property of the index %v of the result is not a string: %v", itemIndex, itemMap["lastName"])
			}
			if _, phoneNumberIsString := itemMap["phoneNumber"].(string); !phoneNumberIsString {
				t.Fatalf("phoneNumber property of the index %v of the result is not a string: %v", itemIndex, itemMap["phoneNumber"])
			}
			if _, usernameIsString := itemMap["username"].(string); !usernameIsString {
				t.Fatalf("username property of the index %v of the result is not a string: %v", itemIndex, itemMap["username"])
			}
		}
	})
	t.Run("custom limit", func(t *testing.T) {
		itemsDefault := []interface{}{}
		getResponse(t, ServerUrl+"/people", &itemsDefault)
		if got, expect := len(itemsDefault), 100; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		itemsCustom := []interface{}{}
		getResponse(t, ServerUrl+"/people?limit=10", &itemsCustom)
		if got, expect := len(itemsCustom), 10; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		if fmt.Sprintf("%#v", itemsCustom[0]) != fmt.Sprintf("%#v", itemsDefault[0]) {
			t.Fatalf("Expected the results to be identical. Expected %#v, got %#v", itemsDefault[0], itemsCustom[0])
		}
	})
	t.Run("limit > max", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/people?max_per_page=2000", &items)
		if got, expect := len(items), 100; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("custom offset", func(t *testing.T) {
		itemsDefault := []interface{}{}
		getResponse(t, ServerUrl+"/people?offset=0", &itemsDefault)

		itemsCustom := []interface{}{}
		getResponse(t, ServerUrl+"/people?offset=10", &itemsCustom)

		if fmt.Sprintf("%#v", itemsCustom[0]) == fmt.Sprintf("%#v", itemsDefault[0]) {
			t.Fatalf("Expected the results to be different. Got %#v", itemsDefault[0])
		}
	})
}
