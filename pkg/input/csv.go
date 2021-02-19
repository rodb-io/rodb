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
	"sync"
	"unsafe"
)

type Csv struct {
	config           *config.CsvInput
	source           source.Source
	logger           *logrus.Logger
	sourceReader     io.ReadSeeker
	sourceReaderLock sync.Mutex
	csvReader        *csv.Reader
	csvReaderBuffer  *bufio.Reader
}

func NewCsv(
	config *config.CsvInput,
	source source.Source,
	log *logrus.Logger,
) (*Csv, error) {
	csvInput := &Csv{
		config:           config,
		source:           source,
		logger:           log,
		sourceReaderLock: sync.Mutex{},
	}

	sourceReader, csvReader, err := csvInput.openSource()
	if err != nil {
		return nil, err
	}

	csvInput.sourceReader = sourceReader
	csvInput.csvReader = csvReader
	csvInput.csvReaderBuffer = getCsvReaderBuffer(csvReader)

	return csvInput, nil
}

func (csvInput *Csv) Get(position record.Position) (record.Record, error) {
	csvInput.sourceReaderLock.Lock()
	defer csvInput.sourceReaderLock.Unlock()

	csvInput.sourceReader.Seek(position, io.SeekStart)
	csvInput.csvReaderBuffer.Reset(csvInput.sourceReader)

	row, err := csvInput.csvReader.Read()
	if err != nil {
		if errors.Is(err, csv.ErrFieldCount) {
			csvInput.logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
			err = nil
		} else {
			return nil, fmt.Errorf("Cannot read csv data: %v", err)
		}
	}

	return record.NewCsv(
		csvInput.config,
		row,
		position,
	), nil
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

		csvReaderBuffer := getCsvReaderBuffer(csvReader)

		for {
			position, err := getCsvReaderOffset(sourceReader, csvReaderBuffer)
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

func getCsvReaderBuffer(csvReader *csv.Reader) *bufio.Reader {
	bufferedReaderField := reflect.ValueOf(csvReader).Elem().FieldByName("r")
	bufferedReaderInterface := reflect.NewAt(
		bufferedReaderField.Type(),
		unsafe.Pointer(bufferedReaderField.UnsafeAddr()),
	).Elem().Interface()
	return bufferedReaderInterface.(*bufio.Reader)
}

func getCsvReaderOffset(reader io.ReadSeeker, csvReaderBuffer *bufio.Reader) (int64, error) {
	offset, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	bufferSize := int64(csvReaderBuffer.Buffered())

	return offset - bufferSize, nil
}
