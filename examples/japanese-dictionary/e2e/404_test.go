package e2e

import (
	"net/http"
	"strings"
	"testing"
)

func Test404(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		httpServerUrl := strings.Replace(ServerUrl, "https://", "http://", 1)
		getErrorResponse(t, httpServerUrl + "/wrong-url", http.StatusNotFound)
	})
	t.Run("https", func(t *testing.T) {
		getErrorResponse(t, ServerUrl + "/wrong-url", http.StatusNotFound)
	})
}
