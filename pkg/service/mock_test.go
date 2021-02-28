package service

import (
	"testing"
)

func TestMockAddRoute(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		mock := NewMock()

		route := &Route{ResponseType: "application/test"}
		mock.AddRoute(route)

		if got, expect := len(mock.routes), 1; got != expect {
			t.Errorf("Expected the server to contain '%v' routes, got '%+v'", expect, got)
		} else if mock.routes[0] != route {
			t.Errorf("Expected the server routes to contain '%+v', got '%+v'", route, mock.routes)
		}
	})
}
