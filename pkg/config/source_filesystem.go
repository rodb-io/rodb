package config

type FilesystemSourceConfig struct{
	Path string
}

func (config *FilesystemSourceConfig) validate() error {
	// The path will be validated at runtime
	return nil
}
