package main

import (
	"rods/pkg/config"
	"rods/pkg/source"
	"rods/pkg/input"
	flag "github.com/spf13/pflag"
	"github.com/sirupsen/logrus"
)

// TODO go fmt

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
		log.Error(err)
		return
	}

	sources, err := source.NewFromConfigs(config.Sources)
	if err != nil {
		log.Errorf("Error initializing sources: %v", err)
		return
	}
	defer source.Close(sources)

	inputs, err := input.NewFromConfigs(config.Inputs, sources)
	if err != nil {
		log.Errorf("Error initializing inputs: %v", err)
		return
	}
	defer input.Close(inputs)

	log.Infof("Inputs: %+v\n", inputs)
}
