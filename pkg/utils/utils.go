package utils

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

func PString(s string) *string {
	return &s
}

func PInt(i int) *int {
	return &i
}

func PFloat(f float64) *float64 {
	return &f
}

func PBool(b bool) *bool {
	return &b
}
