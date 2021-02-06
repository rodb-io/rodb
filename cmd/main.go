package main

import (
	"io/ioutil"

	"rods/pkg/config"

	flag "github.com/spf13/pflag"
	"github.com/sirupsen/logrus"
)

// TODO go fmt
// TODO extract config and log creator in other functions

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

	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Cannot read config file %v: %v", *configPath, err)
	}

	config, err := config.NewConfigFromYaml(configData, log)
	if err != nil {
		log.Fatalf("Cannot parse config file %v: %v", *configPath, err)
	}

	log.Infof("Config: %+v\n", config)
}
