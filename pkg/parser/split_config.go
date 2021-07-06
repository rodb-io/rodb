package parser

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
)

type SplitConfig struct {
	Name              string  `yaml:"name"`
	Type              string  `yaml:"type"`
	Delimiter         *string `yaml:"delimiter"`
	DelimiterIsRegexp *bool   `yaml:"delimiterIsRegexp"`
	Parser            string  `yaml:"parser"`
	Logger            *logrus.Entry
	DelimiterRegexp   *regexp.Regexp
}

func (config *SplitConfig) GetName() string {
	return config.Name
}

func (config *SplitConfig) IsDelimiterARegexp() bool {
	return config.DelimiterIsRegexp != nil && *config.DelimiterIsRegexp
}

func (config *SplitConfig) GetDelimiter() string {
	// Purposefully not checking the pointer because we want
	// to panic if it's nil since the field is required and
	// checked at validation time
	return *config.Delimiter
}

func (config *SplitConfig) Validate(parsers map[string]Parser, log *logrus.Entry) error {
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

	_, parserExists := parsers[config.Parser]
	if !parserExists {
		return fmt.Errorf("Parser '%v' not found in parsers list.", config.Parser)
	}

	return nil
}

func (config *SplitConfig) Primitive() bool {
	return false
}
