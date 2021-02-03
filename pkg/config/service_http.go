package config

type HttpServiceConfig struct{
	Port uint16
}

func (config *HttpServiceConfig) validate() error {
	return nil
}
