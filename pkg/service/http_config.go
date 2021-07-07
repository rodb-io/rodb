package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"rodb.io/pkg/output"
	"rodb.io/pkg/util"
)

type HttpConfig struct {
	Name       string            `yaml:"name"`
	Type       string            `yaml:"type"`
	Http       *HttpHttpConfig   `yaml:"http"`
	Https      *HttpHttpsConfig  `yaml:"https"`
	ErrorsType string            `yaml:"errorsType"`
	Routes     []HttpRouteConfig `yaml:"routes"`
	Logger     *logrus.Entry
}

type HttpHttpConfig struct {
	Listen string `yaml:"listen"`
}

type HttpHttpsConfig struct {
	Listen          string `yaml:"listen"`
	CertificatePath string `yaml:"certificatePath"`
	PrivateKeyPath  string `yaml:"privateKeyPath"`
}

type HttpRouteConfig struct {
	Output string `yaml:"output"`
	Path   string `yaml:"path"`
}

func (config *HttpConfig) GetName() string {
	return config.Name
}

func (config *HttpConfig) Validate(outputs map[string]output.Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("http.name is required")
	}

	if config.ErrorsType == "" {
		log.Debugf("http.errorsType is not set. Defaulting to application/json")
		config.ErrorsType = "application/json"
	}

	if config.Https == nil && config.Http == nil {
		return errors.New("At least one of the http or https property is required.")
	}
	if config.Http != nil {
		if err := config.Http.Validate(log); err != nil {
			return fmt.Errorf("http.%w", err)
		}
	}
	if config.Https != nil {
		if err := config.Https.Validate(log); err != nil {
			return fmt.Errorf("https.%w", err)
		}
	}

	alreadyExistingPaths := make(map[string]bool)
	for i, routeConfig := range config.Routes {
		if err := routeConfig.Validate(outputs, log); err != nil {
			return fmt.Errorf("http.route[%v].%w", i, err)
		}

		if _, alreadyExists := alreadyExistingPaths[routeConfig.Path]; alreadyExists {
			return fmt.Errorf("http.output[%v]: Duplicate path '%v' in array.", i, routeConfig.Path)
		}
		alreadyExistingPaths[routeConfig.Path] = true
	}

	if !util.IsInArray(config.ErrorsType, []string{
		"application/json",
	}) {
		return errors.New("http.errorsType: type '" + config.ErrorsType + "' is not supported.")
	}

	return nil
}

func (config *HttpHttpConfig) Validate(log *logrus.Entry) error {
	if config.Listen == "" {
		config.Listen = "127.0.0.1:0"
	}

	return nil
}

func (config *HttpHttpsConfig) Validate(log *logrus.Entry) error {
	if config.Listen == "" {
		config.Listen = "127.0.0.1:0"
	}

	if config.CertificatePath == "" {
		return errors.New("certificatePath: This field is required")
	}
	certificateFileInfo, err := os.Stat(config.CertificatePath)
	if os.IsNotExist(err) {
		return errors.New("certificatePath: The file '" + config.CertificatePath + "' does not exist")
	}
	if certificateFileInfo.IsDir() {
		return errors.New("certificatePath: The path '" + config.CertificatePath + "' is not a file")
	}

	if config.PrivateKeyPath == "" {
		return errors.New("privateKeyFile: This field is required")
	}
	privateKeyFile, err := os.Stat(config.PrivateKeyPath)
	if os.IsNotExist(err) {
		return errors.New("privateKeyFile: The file '" + config.PrivateKeyPath + "' does not exist")
	}
	if privateKeyFile.IsDir() {
		return errors.New("privateKeyFile: The path '" + config.PrivateKeyPath + "' is not a file")
	}

	return nil
}

func (config *HttpRouteConfig) Validate(outputs map[string]output.Config, log *logrus.Entry) error {
	if config.Output == "" {
		return fmt.Errorf("output is empty. This field is required")
	}

	if config.Path == "" {
		return fmt.Errorf("path is empty. This field is required")
	}

	_, outputExists := outputs[config.Output]
	if !outputExists {
		return fmt.Errorf("output '%v' not found in outputs list.", config.Output)
	}

	return nil
}
