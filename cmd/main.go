package main

import (
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
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

	parsers, err := parser.NewFromConfigs(config.Parsers, log)
	if err != nil {
		log.Errorf("Error initializing parsers: %v", err)
		return
	}
	defer parser.Close(parsers)

	sources, err := source.NewFromConfigs(config.Sources, log)
	if err != nil {
		log.Errorf("Error initializing sources: %v", err)
		return
	}
	defer source.Close(sources)

	inputs, err := input.NewFromConfigs(config.Inputs, sources, parsers, log)
	if err != nil {
		log.Errorf("Error initializing inputs: %v", err)
		return
	}
	defer input.Close(inputs)

	indexes, err := index.NewFromConfigs(config.Indexes, inputs, log)
	if err != nil {
		log.Errorf("Error initializing indexes: %v", err)
		return
	}
	defer index.Close(indexes)

	services, err := service.NewFromConfigs(config.Services, log)
	if err != nil {
		log.Errorf("Error initializing services: %v", err)
		return
	}
	defer service.Close(services)

	outputs, err := output.NewFromConfigs(config.Outputs, indexes, services, parsers, log)
	if err != nil {
		log.Errorf("Error initializing outputs: %v", err)
		return
	}
	defer output.Close(outputs)

	err = service.Wait(services)
	if err != nil {
		log.Error(err)
		return
	}
}
