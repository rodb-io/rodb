package config

type MemoryMapUniqueIndexConfig struct{
	Input string
	Column string
}

func (config *MemoryMapUniqueIndexConfig) validate() error {
	// The input and column will be validated at runtime
	return nil
}
