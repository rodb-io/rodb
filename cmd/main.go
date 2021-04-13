package main

import (
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
	"rodb.io/pkg/config"
	"rodb.io/pkg/index"
	"rodb.io/pkg/input"
	"rodb.io/pkg/output"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/service"
)

func main() {
	verbose := flag.BoolP("verbose", "v", false, "Enable verbose console output")
	configPath := flag.StringP("config", "c", "rodb.yaml", "Path to the configuration file")
	flag.Parse()

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if *verbose {
		log.SetLevel(logrus.TraceLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	config, err := config.NewConfigFromYamlFile(*configPath, log)
	if err != nil {
		log.Errorf("Error initializing config: %v", err)
		return
	}

	parsers, err := parser.NewFromConfigs(config.Parsers)
	if err != nil {
		log.Errorf("Error initializing parsers: %v", err)
		return
	}
	defer (func() {
		err := parser.Close(parsers)
		if err != nil {
			log.Errorf("Error closing parsers: %v", err)
		}
	})()

	inputs, err := input.NewFromConfigs(config.Inputs, parsers)
	if err != nil {
		log.Errorf("Error initializing inputs: %v", err)
		return
	}
	defer (func() {
		err := input.Close(inputs)
		if err != nil {
			log.Errorf("Error closing inputs: %v", err)
		}
	})()

	indexes, err := index.NewFromConfigs(config.Indexes, inputs)
	if err != nil {
		log.Errorf("Error initializing indexes: %v", err)
		return
	}
	defer (func() {
		err := index.Close(indexes)
		if err != nil {
			log.Errorf("Error closing inputs: %v", err)
		}
	})()

	outputs, err := output.NewFromConfigs(config.Outputs, inputs, indexes, parsers)
	if err != nil {
		log.Errorf("Error initializing outputs: %v", err)
		return
	}
	defer (func() {
		err := output.Close(outputs)
		if err != nil {
			log.Errorf("Error closing outputs: %v", err)
		}
	})()

	services, err := service.NewFromConfigs(config.Services, outputs, log)
	if err != nil {
		log.Errorf("Error initializing services: %v", err)
		return
	}
	go (func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)
		signal := <-signals
		log.Printf("Received signal '%v'. Shutting down...", signal.String())

		err := service.Close(services)
		if err != nil {
			log.Errorf("Error closing services: %v", err)
		}
	})()

	err = service.Wait(services)
	if err != nil {
		log.Error(err)
		return
	}
}
