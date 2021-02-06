package main

import (
	"io/ioutil"

	"rods/pkg/config"

	flag "github.com/spf13/pflag"
	"github.com/sirupsen/logrus"
)

func main() {
	verbose := flag.BoolP("verbose", "v", false, "Enable verbose console output")
	configPath := flag.StringP("config", "c", "rods.yaml", "Path to the configuration file")
	flag.Parse()

	log := logrus.New()
	if *verbose {
		log.SetLevel(logrus.TraceLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Cannot read config file %v: %v", *configPath, err)
		return
	}

	config, err := config.NewConfigFromYaml(configData, log)
	if err != nil {
		log.Fatalf("Cannot parse config file %v: %v", *configPath, err)
		return
	}

	log.Infof("Config: %+v\n", config)
}
// TODO unit test for pkg/config/utils.go
// TODO replace environment variables
