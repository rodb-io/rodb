package input

import (
	"fmt"
	"rods/pkg/config"
	"rods/pkg/source"
)

type Csv struct{
	config *config.CsvInputConfig
	source source.Source
}

func NewCsv(
	config *config.CsvInputConfig,
	sources source.SourceList,
) (*Csv, error) {
	source, sourceExists := sources[config.Source]
	if !sourceExists {
		return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Source)
	}

	return &Csv{
		config: config,
		source: source,
	}, nil
}
