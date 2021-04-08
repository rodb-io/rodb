package util

import (
	"net"
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

func TestGetAddress(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 123}), "127.0.0.1:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 123}), "1.2.3.4:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(100, 0, 0, 0), Port: 123}), "100.0.0.0:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
		if got, expect := GetAddress(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 123}), "127.0.0.1:123"; got != expect {
			t.Errorf("Expected to get '%v', got '%v'.", expect, got)
		}
	})
}
