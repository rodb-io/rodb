package pkg

import (
	yaml "gopkg.in/yaml.v2"
)

type Config struct{
	Sources map[string]SourceConfig
	Inputs map[string]InputConfig
	Indexes map[string]IndexConfig
	Services map[string]ServiceConfig
	Outputs map[string]OutputConfig
}

type SourceConfig struct{
	Filesystem *FilesystemSourceConfig
}
type FilesystemSourceConfig struct{
	Path string
}

type InputConfig struct{
	Csv *CsvInputConfig
}
type CsvInputConfig struct{
	Source string
	IgnoreFirstRow string
	Columns []CsvInputColumnConfig
}
type CsvInputColumnConfig struct{
	Name string
	Type *string
	IgnoreCharacters *string
	DecimalSeparator *string
	TrueValues []string
	FalseValues []string
}

type IndexConfig struct{
	MemoryMapUnique *MemoryMapUniqueIndexConfig
	MemoryMapMultiple *MemoryMapMultipleIndexConfig
}
type MemoryMapUniqueIndexConfig struct{
	Input string
	Column string
}
type MemoryMapMultipleIndexConfig struct{
	Input string
	Column string
}

type ServiceConfig struct{
	Http *HttpServiceConfig
}
type HttpServiceConfig struct{
	Port uint16
}

type OutputConfig struct{
	GraphQL *GraphQLOutputConfig
	JsonArray *JsonArrayOutputConfig
	JsonObject *JsonObjectOutputConfig
}
type GraphQLOutputConfig struct{
	Service string
	Endpoint string
}
type JsonObjectOutputConfig struct{
	Service string
	Endpoint string
	Pattern string
	Index string
}
type JsonArrayOutputConfig struct{
	Service string
	Endpoint string
	Limit *JsonArrayOutputLimitConfig
	Offset *JsonArrayOutputOffsetConfig
	Search []JsonArrayOutputSearchConfig
}

type JsonArrayOutputLimitConfig struct{
	Default *uint
	Max *uint
	Param *string
}
type JsonArrayOutputOffsetConfig struct{
	Param string
}
type JsonArrayOutputSearchConfig struct{
	Param string
	Index string
}

func NewConfigFromYaml(yamlConfig []byte) (*Config, error) {
	config := &Config{}
	err := yaml.UnmarshalStrict(yamlConfig, config)
	if err != nil {
		return nil, err
	}

	return config, err
}
