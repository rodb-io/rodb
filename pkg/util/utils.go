package util

import (
	"strings"
)

func RemoveCharacters(value string, charactersToRemove string) string {
	for _, c := range charactersToRemove {
		value = strings.ReplaceAll(value, string(c), "")
	}

	return value
}

func IsInArray(value string, array []string) bool {
	for _, arrayElement := range array {
		if arrayElement == value {
			return true
		}
	}

	return false
}
