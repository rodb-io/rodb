package input

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"reflect"
	"rods/pkg/config"
	"rods/pkg/record"
	"rods/pkg/source"
	"unsafe"
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

func getCsvReaderOffset(reader io.ReadSeeker, csvReader *csv.Reader) (int64, error) {
	offset, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	// Getting the underlying instance of bufio in a dirty way
	bufferedReaderField := reflect.ValueOf(csvReader).Elem().FieldByName("r")
	bufferedReaderInterface := reflect.NewAt(
		bufferedReaderField.Type(),
		unsafe.Pointer(bufferedReaderField.UnsafeAddr()),
	).Elem().Interface()
	bufferedReader := bufferedReaderInterface.(*bufio.Reader)

	bufferSize := int64(bufferedReader.Buffered())

	return offset - bufferSize, nil
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
			position, err := getCsvReaderOffset(sourceReader, csvReader)
			if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read csv position: %v", err)}
			}

			row, err := csvReader.Read()
			if err == io.EOF {
				break
			} else if errors.Is(err, csv.ErrFieldCount) {
				csvInput.logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
			} else if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read csv data: %v", err)}
				return
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
