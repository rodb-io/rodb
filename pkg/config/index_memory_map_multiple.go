package config

type MemoryMapMultipleIndexConfig struct{
	Input string
	Column string
}

func (config *MemoryMapMultipleIndexConfig) validate() error {
	// The input and column will be validated at runtime
	return nil
}
