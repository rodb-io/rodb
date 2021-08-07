package e2e

import (
	"crypto/tls"
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
