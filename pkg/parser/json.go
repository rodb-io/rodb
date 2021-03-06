package parser

import (
	jsonPackage "encoding/json"
)

type Json struct {
	config *JsonConfig
}

func NewJson(
	config *JsonConfig,
) *Json {
	return &Json{
		config: config,
	}
}

func (json *Json) Name() string {
	return json.config.Name
}

func (json *Json) Primitive() bool {
	return json.config.Primitive()
}

func (json *Json) GetRegexpPattern() string {
	return "" // Not matchable because it's not a primitive
}

func (json *Json) Parse(value string) (interface{}, error) {
	var data interface{}
	err := jsonPackage.Unmarshal([]byte(value), &data)
	return data, err
}
