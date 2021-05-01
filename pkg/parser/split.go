package parser

import (
	"fmt"
	"rodb.io/pkg/config"
	"strings"
)

type Split struct {
	config  *config.SplitParser
	parsers List
}

func NewSplit(
	config *config.SplitParser,
	parsers List,
) *Split {
	return &Split{
		config:  config,
		parsers: parsers,
	}
}

func (split *Split) Name() string {
	return split.config.Name
}

func (split *Split) Primitive() bool {
	return split.config.Primitive()
}

func (split *Split) GetRegexpPattern() string {
	return "" // Not matchable because it's not a primitive
}

func (split *Split) Parse(value string) (interface{}, error) {
	var values []string
	if split.config.IsDelimiterARegexp() {
		values = split.config.DelimiterRegexp.Split(value, -1)
	} else {
		values = strings.Split(value, split.config.GetDelimiter())
	}

	// Need to find this at runtime, because the required parser
	// may not be created before the current one
	parser, parserExists := split.parsers[split.config.Parser]
	if !parserExists {
		return nil, fmt.Errorf("The parser %v was not found in parser list", split.config.Parser)
	}

	parsedValues := make([]interface{}, len(values))
	for i := 0; i < len(values); i++ {
		value, err := parser.Parse(values[i])
		if err != nil {
			return nil, err
		}
		parsedValues[i] = value
	}

	return parsedValues, nil
}
