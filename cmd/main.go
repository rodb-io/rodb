package main

import (
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"rods/pkg/config"
	"rods/pkg/input"
	"rods/pkg/source"
)

// TODO create a non-indexed index, have it as default value if no index is specified by the output search/filter

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

	sources, err := source.NewFromConfigs(config.Sources, log)
	if err != nil {
		log.Errorf("Error initializing sources: %v", err)
		return
	}
	defer source.Close(sources)

	inputs, err := input.NewFromConfigs(config.Inputs, sources, log)
	if err != nil {
		log.Errorf("Error initializing inputs: %v", err)
		return
	}
	defer input.Close(inputs)

	log.Infof("Inputs: %+v\n", inputs)
}
