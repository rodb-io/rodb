package config

type JsonArrayOutputConfig struct{
	Service string
	Endpoint string
	Limit JsonArrayOutputLimitConfig
	Offset JsonArrayOutputOffsetConfig
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

func (config *JsonArrayOutputConfig) validate() error {
	return nil
}
