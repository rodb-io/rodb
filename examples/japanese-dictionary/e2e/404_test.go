package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func Test404(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		httpServerUrl := strings.Replace(ServerUrl, "https://", "http://", 1)
		response, err := getClient().Get(httpServerUrl + "/wrong-url")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if got, expect := response.StatusCode, http.StatusNotFound; got != expect {
			t.Fatalf("Got return status %v, expected %v", got, expect)
		}

		body := map[string]string{}
		jsonDecoder := json.NewDecoder(response.Body)
		if err := jsonDecoder.Decode(&body); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if _, hasError := body["error"]; !hasError {
			t.Fatalf("Expected an error property in the json body, got %v", body)
		}
	})
	t.Run("https", func(t *testing.T) {
		response, err := getClient().Get(ServerUrl + "/wrong-url")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if got, expect := response.StatusCode, http.StatusNotFound; got != expect {
			t.Fatalf("Got return status %v, expected %v", got, expect)
		}

		body := map[string]string{}
		jsonDecoder := json.NewDecoder(response.Body)
		if err := jsonDecoder.Decode(&body); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if _, hasError := body["error"]; !hasError {
			t.Fatalf("Expected an error property in the json body, got %v", body)
		}
	})
}
