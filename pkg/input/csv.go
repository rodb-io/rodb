package input

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"rods/pkg/config"
	"rods/pkg/record"
	"rods/pkg/source"
)

type Csv struct {
	config       *config.CsvInput
	source       source.Source
	logger       *logrus.Logger
	sourceReader io.ReadSeeker
	csvReader    *csv.Reader
}

func NewCsv(
	config *config.CsvInput,
	source source.Source,
	log *logrus.Logger,
) (*Csv, error) {
	csvInput := &Csv{
		config: config,
		source: source,
		logger: log,
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
	csvReader.ReuseRecord = false

	return sourceReader, csvReader, nil
}

func (csvInput *Csv) IterateAll() <-chan IterateAllResult {
	channel := make(chan IterateAllResult)

	go func() {
		defer close(channel)

		sourceReader, csvReader, err := csvInput.openSource()
		if err != nil {
			channel <- IterateAllResult{Error: err}
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
			} else if errors.Is(err, csv.ErrFieldCount) {
				csvInput.logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
			} else if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read csv data: %v", err)}
				return
			}

			position, err := sourceReader.Seek(0, io.SeekCurrent)
			if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read csv position: %v", err)}
			}

			channel <- IterateAllResult{
				Record: record.NewCsv(
					csvInput.config,
					row,
					position,
				),
			}
		}
	}()

	return channel
}

func (csvInput *Csv) Close() error {
	return csvInput.source.CloseReader(csvInput.sourceReader)
}
