package util

import (
	"errors"
)

func GetTypeFromConfigUnmarshaler(unmarshal func(interface{}) error) (string, error) {
	asMap := map[string]interface{}{}
	err := unmarshal(asMap)
	if err != nil {
		return "", err
	}

	objectType, objectTypeExists := asMap["type"]
	if !objectTypeExists {
		return "", errors.New("The 'type' attribute is required.")
	}

	objectTypeString, objectTypeIsString := objectType.(string)
	if !objectTypeIsString {
		return "", errors.New("The 'type' attribute must be a string.")
	}
	if objectTypeString == "" {
		return "", errors.New("The 'type' attribute cannot be empty.")
	}

	return objectTypeString, nil
}
