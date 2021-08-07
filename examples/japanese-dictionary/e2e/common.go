package e2e

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const ServerUrl = "https://japanese-dictionary"

func TestMain(m *testing.M) {
	request, err := http.NewRequest("GET", ServerUrl, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var response *http.Response
	for i := 0; i < 30; i++ {
		if response, err = getClient().Do(request); err != nil {
			time.Sleep(250 * time.Millisecond) // and continue
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

func getClient() *http.Client {
	return &http.Client{
		Timeout: 250 * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func getListResponse(t *testing.T, url string) []interface{} {
	response, err := getClient().Get(url)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got, expect := response.StatusCode, http.StatusOK; got != expect {
		t.Fatalf("Got return status %v, expected %v", got, expect)
	}

	body := []interface{}{}
	jsonDecoder := json.NewDecoder(response.Body)
	if err := jsonDecoder.Decode(&body); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	return body
}

func getErrorResponse(t *testing.T, url string, expectStatus int) map[string]string {
	response, err := getClient().Get(url)
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
