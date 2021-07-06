package parser

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
	"io/ioutil"
	"strings"
)

type String struct {
	config  *StringConfig
	decoder *encoding.Decoder
}

func NewString(
	config *StringConfig,
) (*String, error) {
	var decoder *encoding.Decoder = nil
	if config.ConvertFromCharset != "" {
		encoding, err := ianaindex.MIME.Encoding(config.ConvertFromCharset)
		if err != nil {
			return nil, err
		}

		decoder = encoding.NewDecoder()
	}

	return &String{
		config:  config,
		decoder: decoder,
	}, nil
}

func (str *String) Name() string {
	return str.config.Name
}

func (str *String) Primitive() bool {
	return str.config.Primitive()
}

func (str *String) GetRegexpPattern() string {
	return ".+"
}

func (str *String) Parse(value string) (interface{}, error) {
	if str.config.ConvertFromCharset != "" {
		rInUTF8 := transform.NewReader(strings.NewReader(value), str.decoder)
		convertedValue, err := ioutil.ReadAll(rInUTF8)
		if err != nil {
			return nil, err
		}

		value = string(convertedValue)
	}

	return value, nil
}
