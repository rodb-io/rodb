package parser

import (
	"fmt"
	"regexp"
	"github.com/rodb-io/rodb/pkg/util"
	"strings"
)

type Boolean struct {
	config *BooleanConfig
}

func NewBoolean(
	config *BooleanConfig,
) *Boolean {
	return &Boolean{
		config: config,
	}
}

func (boolean *Boolean) Name() string {
	return boolean.config.Name
}

func (boolean *Boolean) Primitive() bool {
	return boolean.config.Primitive()
}

func (boolean *Boolean) GetRegexpPattern() string {
	values := make([]string, 0, len(boolean.config.TrueValues)+len(boolean.config.FalseValues))
	for _, value := range boolean.config.TrueValues {
		values = append(values, regexp.QuoteMeta(value))
	}
	for _, value := range boolean.config.FalseValues {
		values = append(values, regexp.QuoteMeta(value))
	}
	return "(" + strings.Join(values, "|") + ")"
}

func (boolean *Boolean) Parse(value string) (interface{}, error) {
	if util.IsInArray(value, boolean.config.TrueValues) {
		return true, nil
	}
	if util.IsInArray(value, boolean.config.FalseValues) {
		return false, nil
	}

	return nil, fmt.Errorf("The value '%v' was found but is neither declared in trueValues or falseValues.", value)
}
