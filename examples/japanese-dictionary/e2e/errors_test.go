package e2e

import (
	"net/http"
	"strings"
	"testing"
)

func TestErrors(t *testing.T) {
	t.Run("404", func(t *testing.T) {
		getErrorResponse(t, ServerUrl + "/wrong-url", http.StatusNotFound)
	})
	t.Run("404 via http", func(t *testing.T) {
		httpServerUrl := strings.Replace(ServerUrl, "https://", "http://", 1)
		getErrorResponse(t, httpServerUrl + "/wrong-url", http.StatusNotFound)
	})
}
