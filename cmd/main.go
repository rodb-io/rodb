package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"rods/pkg"

	flag "github.com/spf13/pflag"
)

func main() {
	configPath := flag.String("config", "rods.yaml", "Path to the configuration file")
	flag.Parse()

	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read config file %v: %v", *configPath, err)
		return
	}

	config, err := pkg.NewConfigFromYaml(configData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse config file %v: %v", *configPath, err)
		return
	}

	fmt.Printf("Config: %+v\n", config)
}
