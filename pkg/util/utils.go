package util

import (
	"net"
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

func GetAddress(address net.Addr) string {
	result := address.String()
	for from, to := range map[string]string{
		"[::]:":    "127.0.0.1:",
		"0.0.0.0:": "127.0.0.1:",
	} {
		if strings.HasPrefix(result, from) {
			result = to + result[len(from):]
		}
	}

	return result
}
