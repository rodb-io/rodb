package record

import (
	"fmt"
	"strconv"
)

func getSubValue(value interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return value, nil
	}

	switch value.(type) {
	case map[string]interface{}:
		mapValue, mapValueExists := value.(map[string]interface{})[path[0]]
		if !mapValueExists {
			// Not having some properties is a common case that
			// should not trigger an error, but get a nil value
			return nil, nil
		}

		return getSubValue(mapValue, path[1:])
	case []interface{}:
		index, err := strconv.Atoi(path[0])
		if err != nil {
			return nil, fmt.Errorf("Cannot get sub-path '%v' because the requested key is '%v', but the value is an array '%#v'", path, path[0], value)
		}

		valueArray := value.([]interface{})
		if index >= len(valueArray) {
			// Not having the required index is a common case that
			// should not trigger an error, but get a nil value
			return nil, nil
		}

		return getSubValue(valueArray[index], path[1:])
	default:
		return nil, fmt.Errorf("Cannot get sub-path '%v' because the value is primitive or unknown: '%#v'", path, value)
	}
}
