package e2e

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	waitForServer(t)
	t.Run("list", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes", &items)
		if got, expect := len(items), 30; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		for itemIndex, item := range items {
			itemMap, isMap := item.(map[string]interface{})
			if !isMap {
				t.Fatalf("Item %v of the result is not an object: %v", itemIndex, item)
			}

			if _, codeIsFloat := itemMap["code"].(float64); !codeIsFloat {
				t.Fatalf("code property of the index %v of the result is not a float64: %v", itemIndex, itemMap["code"])
			}
			if _, hasSubdivisionIsBoolean := itemMap["hasSubdivision"].(bool); !hasSubdivisionIsBoolean {
				t.Fatalf("hasSubdivision property of the index %v of the result is not a bool: %v", itemIndex, itemMap["hasSubdivision"])
			}
			if _, municipalityIsString := itemMap["municipality"].(string); !municipalityIsString {
				t.Fatalf("municipality property of the index %v of the result is not a string: %v", itemIndex, itemMap["municipality"])
			}
			if _, municipalityKanaIsString := itemMap["municipalityKana"].(string); !municipalityKanaIsString {
				t.Fatalf("municipalityKana property of the index %v of the result is not a string: %v", itemIndex, itemMap["municipalityKana"])
			}
			if _, oldZipCodeIsFloat := itemMap["oldZipCode"].(float64); !oldZipCodeIsFloat {
				t.Fatalf("oldZipCode property of the index %v of the result is not a float64: %v", itemIndex, itemMap["oldZipCode"])
			}
			if _, prefectureIsString := itemMap["prefecture"].(string); !prefectureIsString {
				t.Fatalf("prefecture property of the index %v of the result is not a string: %v", itemIndex, itemMap["prefecture"])
			}
			if _, prefectureKanaIsString := itemMap["prefectureKana"].(string); !prefectureKanaIsString {
				t.Fatalf("prefectureKana property of the index %v of the result is not a string: %v", itemIndex, itemMap["prefectureKana"])
			}
			if _, reasonForUpdateIdIsFloat := itemMap["reasonForUpdateId"].(float64); !reasonForUpdateIdIsFloat {
				t.Fatalf("reasonForUpdateId property of the index %v of the result is not a float64: %v", itemIndex, itemMap["reasonForUpdateId"])
			}
			if _, streetNumberAssignedPerKanaIsBoolean := itemMap["streetNumberAssignedPerKana"].(bool); !streetNumberAssignedPerKanaIsBoolean {
				t.Fatalf("streetNumberAssignedPerKana property of the index %v of the result is not a bool: %v", itemIndex, itemMap["streetNumberAssignedPerKana"])
			}
			if _, townIsString := itemMap["town"].(string); !townIsString {
				t.Fatalf("town property of the index %v of the result is not a string: %v", itemIndex, itemMap["town"])
			}
			if _, townHasMultipleZipCodesIsBoolean := itemMap["townHasMultipleZipCodes"].(bool); !townHasMultipleZipCodesIsBoolean {
				t.Fatalf("townHasMultipleZipCodes property of the index %v of the result is not a bool: %v", itemIndex, itemMap["townHasMultipleZipCodes"])
			}
			if _, townKanaIsString := itemMap["townKana"].(string); !townKanaIsString {
				t.Fatalf("townKana property of the index %v of the result is not a string: %v", itemIndex, itemMap["townKana"])
			}
			if _, updatedIdIsFloat := itemMap["updatedId"].(float64); !updatedIdIsFloat {
				t.Fatalf("updatedId property of the index %v of the result is not a float64: %v", itemIndex, itemMap["updatedId"])
			}
			if _, zipCodeIsString := itemMap["zipCode"].(string); !zipCodeIsString {
				t.Fatalf("zipCode property of the index %v of the result is not a string: %v", itemIndex, itemMap["zipCode"])
			}
			if _, zipCodeHasMultipleTownsIsBoolean := itemMap["zipCodeHasMultipleTowns"].(bool); !zipCodeHasMultipleTownsIsBoolean {
				t.Fatalf("zipCodeHasMultipleTowns property of the index %v of the result is not a bool: %v", itemIndex, itemMap["zipCodeHasMultipleTowns"])
			}

			reasonsForUpdate, reasonsForUpdateIsMap := itemMap["reasonsForUpdate"].(map[string]interface{})
			if !reasonsForUpdateIsMap {
				t.Fatalf("reasonsForUpdate property of the index %v of the result is not a map: %v", itemIndex, itemMap["reasonsForUpdate"])
			}
			if _, reasonsForUpdateIdIsFloat := reasonsForUpdate["id"].(float64); !reasonsForUpdateIdIsFloat {
				t.Fatalf("reasonsForUpdate.id property of the index %v of the result is not a float64: %v", itemIndex, reasonsForUpdate["id"])
			}
			if _, reasonsForUpdateNameIsString := reasonsForUpdate["name"].(string); !reasonsForUpdateNameIsString {
				t.Fatalf("reasonsForUpdate.name property of the index %v of the result is not a string: %v", itemIndex, reasonsForUpdate["name"])
			}

			updated, updatedIsMap := itemMap["updated"].(map[string]interface{})
			if !updatedIsMap {
				t.Fatalf("updated property of the index %v of the result is not a map: %v", itemIndex, itemMap["updated"])
			}
			if _, updatedIdIsFloat := updated["id"].(float64); !updatedIdIsFloat {
				t.Fatalf("updated.id property of the index %v of the result is not a float64: %v", itemIndex, updated["id"])
			}
			if _, updatedNameIsString := updated["name"].(string); !updatedNameIsString {
				t.Fatalf("updated.name property of the index %v of the result is not a string: %v", itemIndex, updated["name"])
			}
		}
	})
	t.Run("custom max_per_page", func(t *testing.T) {
		itemsDefault := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes", &itemsDefault)
		if got, expect := len(itemsDefault), 30; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		itemsCustom := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?max_per_page=10", &itemsCustom)
		if got, expect := len(itemsCustom), 10; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}

		if fmt.Sprintf("%#v", itemsCustom[0]) != fmt.Sprintf("%#v", itemsDefault[0]) {
			t.Fatalf("Expected the results to be identical. Expected %#v, got %#v", itemsDefault[0], itemsCustom[0])
		}
	})
	t.Run("max_per_page > max", func(t *testing.T) {
		items := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?max_per_page=150", &items)
		if got, expect := len(items), 100; got != expect {
			t.Fatalf("Got %v items, expected %v", got, expect)
		}
	})
	t.Run("custom offset_from", func(t *testing.T) {
		itemsDefault := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?offset_from=0", &itemsDefault)

		itemsCustom := []interface{}{}
		getResponse(t, ServerUrl+"/zip-codes?offset_from=10", &itemsCustom)

		if fmt.Sprintf("%#v", itemsCustom[0]) == fmt.Sprintf("%#v", itemsDefault[0]) {
			t.Fatalf("Expected the results to be different. Got %#v", itemsDefault[0])
		}
	})
}
