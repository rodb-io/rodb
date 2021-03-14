package main

import (
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
	"rods/pkg/config"
	"rods/pkg/index"
	"rods/pkg/input"
	"rods/pkg/output"
	"rods/pkg/parser"
	"rods/pkg/service"
	"rods/pkg/source"
)

func main() {
	verbose := flag.BoolP("verbose", "v", false, "Enable verbose console output")
	configPath := flag.StringP("config", "c", "rods.yaml", "Path to the configuration file")
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

	sources, err := source.NewFromConfigs(config.Sources)
	if err != nil {
		log.Errorf("Error initializing sources: %v", err)
		return
	}
	defer (func() {
		err := source.Close(sources)
		if err != nil {
			log.Errorf("Error closing sources: %v", err)
		}
	})()

	inputs, err := input.NewFromConfigs(config.Inputs, sources, parsers)
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

	services, err := service.NewFromConfigs(config.Services, log)
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

	outputs, err := output.NewFromConfigs(config.Outputs, indexes, services, parsers)
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

	err = service.Wait(services)
	if err != nil {
		log.Error(err)
		return
	}
}
