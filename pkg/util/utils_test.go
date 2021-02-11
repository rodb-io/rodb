package util

import (
	"testing"
)

func TestRemoveCharacters(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		if got, expect := RemoveCharacters("abcdef", "db"), "acef"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("nothing to replace", func(t *testing.T) {
		if got, expect := RemoveCharacters("abcdef", "ghi"), "abcdef"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
	t.Run("unicode character", func(t *testing.T) {
		if got, expect := RemoveCharacters("あいうえお", "うお"), "あいえ"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'", expect, got)
		}
	})
}

func TestIsInArray(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if result := IsInArray("string", []string{"a", "string"}); !result {
			t.Fail()
		}
	})
	t.Run("invalid", func(t *testing.T) {
		if result := IsInArray("invalid", []string{"string"}); result {
			t.Fail()
		}
	})
	t.Run("empty value", func(t *testing.T) {
		if result := IsInArray("", []string{"string"}); result {
			t.Fail()
		}
	})
	t.Run("empty array", func(t *testing.T) {
		if result := IsInArray("string", []string{}); result {
			t.Fail()
		}
	})
}
