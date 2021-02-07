package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"rods/pkg/config"
	"rods/pkg/source"
)

type Csv struct{
	config *config.CsvInputConfig
	source source.Source
	sourceReader io.ReadSeeker
	csvReader *csv.Reader
}

func NewCsv(
	config *config.CsvInputConfig,
	sources source.SourceList,
) (*Csv, error) {
	source, sourceExists := sources[config.Source]
	if !sourceExists {
		return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Source)
	}

	sourceReader, err := source.Open(config.Path)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(sourceReader)
	csvReader.Comma = []rune(config.Delimiter)[0]
	csvReader.FieldsPerRecord = len(config.Columns)

	return &Csv{
		config: config,
		source: source,
		sourceReader: sourceReader,
		csvReader: csvReader,
	}, nil
}

func (csv *Csv) Close() error {
	return csv.source.CloseReader(csv.sourceReader)
}
