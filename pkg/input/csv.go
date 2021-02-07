package input

import (
	"fmt"
	"io"
	"rods/pkg/config"
	"rods/pkg/source"
)

type Csv struct{
	config *config.CsvInputConfig
	source source.Source
	reader io.ReadSeeker
}

func NewCsv(
	config *config.CsvInputConfig,
	sources source.SourceList,
) (*Csv, error) {
	source, sourceExists := sources[config.Source]
	if !sourceExists {
		return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Source)
	}

	reader, err := source.Open(config.Path)
	if err != nil {
		return nil, err
	}

	return &Csv{
		config: config,
		source: source,
		reader: reader,
	}, nil
}

func (csv *Csv) Close() error {
	return csv.source.CloseReader(csv.reader)
}
