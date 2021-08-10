package e2e

import (
	"strings"
	"testing"
)

func TestParameters(t *testing.T) {
	t.Run("person found with search", func(t *testing.T) {
		search := "John"
		items := []interface{}{}
		getResponse(t, ServerUrl+"/people?search="+search, &items)

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			firstName, firstNameIsString := itemMap["firstName"].(string)
			if !firstNameIsString {
				t.Fatalf("Expected the firstName property to be a string, got: %#v", item)
			}
			lastName, lastNameIsString := itemMap["lastName"].(string)
			if !lastNameIsString {
				t.Fatalf("Expected the lastName property to be a string, got: %#v", item)
			}

			if !strings.Contains(firstName, search) && !strings.Contains(lastName, search) {
				t.Fatalf("Expected either the firstName of lastName property to contain '%v', got: %#v", search, item)
			}
		}
	})
	t.Run("get item", func(t *testing.T) {
		item := map[string]interface{}{}
		getResponse(t, ServerUrl+"/people/2", &item)

		if got, expect := item["firstName"], "Aubree"; got != expect {
			t.Fatalf("firstName property of the result does not match the input parameter: %#v", item)
		}
		if got, expect := item["lastName"], "Moen"; got != expect {
			t.Fatalf("lastName property of the result does not match the input parameter: %#v", item)
		}
	})
}
