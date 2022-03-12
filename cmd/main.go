package main

import (
	"fmt"
	"github.com/rodb-io/rodb/pkg/config"
	"github.com/rodb-io/rodb/pkg/index"
	"github.com/rodb-io/rodb/pkg/input"
	"github.com/rodb-io/rodb/pkg/output"
	"github.com/rodb-io/rodb/pkg/parser"
	"github.com/rodb-io/rodb/pkg/service"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
)

func main() {
	logLevelString := flag.StringP(
		"loglevel",
		"l",
		logrus.InfoLevel.String(),
		"Changes the logging level.\nSupported values: panic, fatal, error, warn[ing], info, debug, trace",
	)
	configPath := flag.StringP("config", "c", "rodb.yaml", "Path to the configuration file")
	flag.Parse()

	logLevel, err := logrus.ParseLevel(*logLevelString)
	if err != nil {
		fmt.Printf("Error: %v", err)
		flag.PrintDefaults()
		os.Exit(1)
		return
	}

	log := logrus.New()
	log.SetLevel(logLevel)
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	config, err := config.NewConfigFromYamlFile(*configPath, log)
	if err != nil {
		log.Errorf("Error initializing config: %v", err)
		os.Exit(1)
		return
	}

	parsers, err := parser.NewFromConfigs(config.Parsers)
	if err != nil {
		log.Errorf("Error initializing parsers: %v", err)
		os.Exit(1)
		return
	}
	defer (func() {
		if err := parser.Close(parsers); err != nil {
			log.Errorf("Error closing parsers: %v", err)
		}
	})()

	inputs, err := input.NewFromConfigs(config.Inputs, parsers)
	if err != nil {
		log.Errorf("Error initializing inputs: %v", err)
		os.Exit(1)
		return
	}
	defer (func() {
		if err := input.Close(inputs); err != nil {
			log.Errorf("Error closing inputs: %v", err)
		}
	})()

	indexes, err := index.NewFromConfigs(config.Indexes, inputs)
	if err != nil {
		log.Errorf("Error initializing indexes: %v", err)
		os.Exit(1)
		return
	}
	defer (func() {
		if err := index.Close(indexes); err != nil {
			log.Errorf("Error closing inputs: %v", err)
		}
	})()

	outputs, err := output.NewFromConfigs(config.Outputs, inputs, indexes, parsers)
	if err != nil {
		log.Errorf("Error initializing outputs: %v", err)
		os.Exit(1)
		return
	}
	defer (func() {
		if err := output.Close(outputs); err != nil {
			log.Errorf("Error closing outputs: %v", err)
		}
	})()

	services, err := service.NewFromConfigs(config.Services, outputs, log)
	if err != nil {
		log.Errorf("Error initializing services: %v", err)
		os.Exit(1)
		return
	}
	go (func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)
		signal := <-signals
		log.Printf("Received signal '%v'. Shutting down...", signal.String())

		if err := service.Close(services); err != nil {
			log.Errorf("Error closing services: %v", err)
		}
	})()

	if err := service.Wait(services); err != nil {
		log.Error(err)
		os.Exit(1)
		return
	}
}
