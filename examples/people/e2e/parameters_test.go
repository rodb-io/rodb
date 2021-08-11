package e2e

import (
	"fmt"
	"strings"
	"testing"
)

func TestParameters(t *testing.T) {
	waitForServer(t)
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
		itemsList := []interface{}{}
		getResponse(t, ServerUrl+"/people", &itemsList)

		randomItem, randomItemIsMap := itemsList[2].(map[string]interface{})
		if !randomItemIsMap {
			t.Fatalf("Expected the item to be a map, got: %#v", itemsList[2])
		}

		id := int(randomItem["id"].(float64))
		if id == 0 {
			t.Fatalf("Expected the object to have an id, got: %#v", randomItem)
		}

		firstName := randomItem["firstName"].(string)
		if firstName == "" {
			t.Fatalf("Expected the object to have a firstName, got: %#v", randomItem)
		}

		lastName := randomItem["lastName"].(string)
		if lastName == "" {
			t.Fatalf("Expected the object to have a lastName, got: %#v", randomItem)
		}

		item := map[string]interface{}{}
		getResponse(t, fmt.Sprintf("%v/people/%v", ServerUrl, id), &item)

		if got, expect := item["firstName"], firstName; got != expect {
			t.Fatalf("firstName property of the result does not match the expected '%v' result: %#v", expect, item)
		}
		if got, expect := item["lastName"], lastName; got != expect {
			t.Fatalf("lastName property of the result does not match the expected '%v' result: %#v", expect, item)
		}
	})
}
