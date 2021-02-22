package service

import (
	"context"
	"io"
	"github.com/sirupsen/logrus"
	"net/http"
	"rods/pkg/config"
	"strconv"
	"sync"
)

type Http struct {
	router *http.ServeMux
	server *http.Server
	waitGroup *sync.WaitGroup
}

func NewHttp(
	config *config.HttpService,
	waitGroup *sync.WaitGroup,
	log *logrus.Logger,
) (*Http, error) {
	router := http.NewServeMux()

	server := &http.Server{
		Addr: ":" + strconv.Itoa(int(config.Port)),
		Handler: router,
	}

	// TODO health check + method to add a route
	router.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		io.WriteString(response, "Hello world!\n")
	})

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Http service: %v", err)
		}
	}()

	return &Http{
		router: router,
		server: server,
		waitGroup: waitGroup,
	}, nil
}

func (http *Http) Close() error {
	err := http.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	http.waitGroup.Wait()
	return nil
}
