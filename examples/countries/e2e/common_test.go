package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const ServerUrl = "http://countries"

func waitForServer(t *testing.T) {
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	request, err := http.NewRequest("GET", ServerUrl, nil)
	if err != nil {
		t.Fatal(err)
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
		t.Fatalf("The server %v did not start before the end of the check period.\n", ServerUrl)
	}
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
