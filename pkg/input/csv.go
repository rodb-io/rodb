package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"rods/pkg/config"
	"rods/pkg/source"
	"github.com/sirupsen/logrus"
)

type Csv struct{
	config *config.CsvInputConfig
	source source.Source
	sourceReader io.ReadSeeker
	csvReader *csv.Reader
	logger *logrus.Logger
}

func NewCsv(
	config *config.CsvInputConfig,
	sources source.SourceList,
	log *logrus.Logger,
) (*Csv, error) {
	source, sourceExists := sources[config.Source]
	if !sourceExists {
		return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Source)
	}

	csvInput := &Csv{
		config: config,
		source: source,
	}

	sourceReader, csvReader, err := csvInput.openSource()
	if err != nil {
		return nil, err
	}

	csvInput.sourceReader = sourceReader
	csvInput.csvReader = csvReader

	return csvInput, nil
}

func (csvInput *Csv) openSource() (io.ReadSeeker, *csv.Reader, error) {
	sourceReader, err := csvInput.source.Open(csvInput.config.Path)
	if err != nil {
		return nil, nil, err
	}

	csvReader := csv.NewReader(sourceReader)
	csvReader.Comma = []rune(csvInput.config.Delimiter)[0]
	csvReader.FieldsPerRecord = len(csvInput.config.Columns)

	return sourceReader, csvReader, nil
}

func (csvInput *Csv) IterateAll() (chan []string, chan error) {
	rowsChannel := make(chan []string)
	errorsChannel := make(chan error)

	go func() {
		defer close(rowsChannel)
		defer close(errorsChannel)

		sourceReader, csvReader, err := csvInput.openSource()
		if err != nil {
			errorsChannel <- err
			return
		}
		defer csvInput.source.CloseReader(sourceReader)

		sourceReader.Seek(0, io.SeekStart)
		if csvInput.config.IgnoreFirstRow {
			_, _ = csvReader.Read()
		}

		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				break
			} else if err == csv.ErrFieldCount {
				csvInput.logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
			} else if err != nil {
				errorsChannel <- fmt.Errorf("Cannot read csv data: %v", err)
				return
			}

			rowsChannel <- row
		}
	}()

	return rowsChannel, errorsChannel
}

func (csvInput *Csv) Close() error {
	return csvInput.source.CloseReader(csvInput.sourceReader)
}
