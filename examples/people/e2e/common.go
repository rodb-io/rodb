package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const ServerUrl = "http://people"

func TestMain(m *testing.M) {
	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	request, err := http.NewRequest("GET", ServerUrl, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var response *http.Response
	for {
		if response, err = client.Do(request); err != nil {
			time.Sleep(500 * time.Millisecond) // and continue
		} else {
			break
		}
	}

	if response == nil {
		fmt.Printf("The server %v did not start before the end of the check period.\n", ServerUrl)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func getResponse(t *testing.T, url string, out interface{}) {
	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got, expect := response.StatusCode, http.StatusOK; got != expect {
		t.Fatalf("Got return status %v, expected %v", got, expect)
	}

	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(out); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func getErrorResponse(t *testing.T, url string, expectStatus int) map[string]string {
	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got := response.StatusCode; got != expectStatus {
		t.Fatalf("Got return status %v, expected %v", got, expectStatus)
	}

	body := map[string]string{}
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&body); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, hasError := body["error"]; !hasError {
		t.Fatalf("Expected an error property in the json body, got %v", body)
	}

	return body
}
