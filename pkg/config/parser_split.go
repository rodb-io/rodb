package config

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
)

type SplitParser struct {
	Name              string  `yaml:"name"`
	Delimiter         *string `yaml:"delimiter"`
	DelimiterIsRegexp *bool   `yaml:"delimiterIsRegexp"`
	Parser            string  `yaml:"parser"`
	Logger            *logrus.Entry
	DelimiterRegexp   *regexp.Regexp
}

func (config *SplitParser) validate(rootConfig *Config, log *logrus.Entry) error {
	config.Logger = log

	if config.Name == "" {
		return errors.New("split.name is required")
	}

	if config.Delimiter == nil {
		return errors.New("split.delimiter is required")
	}

	if config.DelimiterIsRegexp == nil {
		falseValue := false
		config.DelimiterIsRegexp = &falseValue
		log.Debug("split.delimiterIsRegexp is not set. Assuming false")
	}

	if *config.DelimiterIsRegexp == true {
		var err error
		config.DelimiterRegexp, err = regexp.Compile(*config.Delimiter)
		if err != nil {
			return fmt.Errorf("split.delimiter: Error while parsing regexp: %w", err)
		}
	}

	_, parserExists := rootConfig.Parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("Parser '%v' not found in parsers list.", config.Parser)
	}

	return nil
}

func (config *SplitParser) Primitive() bool {
	return false
}
