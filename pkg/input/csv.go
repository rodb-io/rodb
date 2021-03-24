package input

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	configModule "rods/pkg/config"
	"rods/pkg/parser"
	"rods/pkg/record"
	"rods/pkg/source"
	"sync"
	"unsafe"
)

type Csv struct {
	config           *configModule.CsvInput
	source           source.Source
	sourceReader     io.ReadSeeker
	sourceReaderLock sync.Mutex
	csvReader        *csv.Reader
	csvReaderBuffer  *bufio.Reader
	columnParsers    []parser.Parser
}

func NewCsv(
	config *configModule.CsvInput,
	sources source.List,
	parsers parser.List,
) (*Csv, error) {
	source, sourceExists := sources[config.Source]
	if !sourceExists {
		return nil, fmt.Errorf("Source '%v' not found in sources list.", config.Source)
	}

	csvInput := &Csv{
		config:           config,
		source:           source,
		sourceReaderLock: sync.Mutex{},
	}

	sourceReader, csvReader, err := csvInput.openSource()
	if err != nil {
		return nil, err
	}
	csvInput.sourceReader = sourceReader
	csvInput.csvReader = csvReader
	csvInput.csvReaderBuffer = getCsvReaderBuffer(csvReader)

	if config.AutodetectColumns {
		err := csvInput.autodetectColumns()
		if err != nil {
			return nil, err
		}
	}

	csvInput.columnParsers = make([]parser.Parser, len(config.Columns))
	for i, column := range config.Columns {
		parser, parserExists := parsers[column.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + column.Parser + "' does not exist")
		}
		csvInput.columnParsers[i] = parser
	}

	return csvInput, nil
}

func (csvInput *Csv) Name() string {
	return csvInput.config.Name
}

func (csvInput *Csv) HasColumn(columnName string) bool {
	for _, column := range csvInput.config.Columns {
		if column.Name == columnName {
			return true
		}
	}

	return false
}

func (csvInput *Csv) Get(position record.Position) (record.Record, error) {
	csvInput.sourceReaderLock.Lock()
	defer csvInput.sourceReaderLock.Unlock()

	csvInput.sourceReader.Seek(position, io.SeekStart)
	csvInput.csvReaderBuffer.Reset(csvInput.sourceReader)

	row, err := csvInput.csvReader.Read()
	if err != nil {
		if errors.Is(err, csv.ErrFieldCount) {
			csvInput.config.Logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
			err = nil
		} else {
			return nil, fmt.Errorf("Cannot read csv data: %w", err)
		}
	}

	return record.NewCsv(
		csvInput.config,
		csvInput.columnParsers,
		row,
		position,
	), nil
}

func (csvInput *Csv) Size() (int64, error) {
	return csvInput.source.Size(csvInput.config.Path)
}

func (csvInput *Csv) autodetectColumns() error {
	firstRow, err := csvInput.csvReader.Read()
	if err != nil {
		return fmt.Errorf("Cannot read csv data: %w", err)
	}

	alreadyExistingNames := make(map[string]bool)
	csvInput.config.Columns = make([]*configModule.CsvInputColumn, len(firstRow))
	csvInput.config.ColumnIndexByName = make(map[string]int)
	for columnIndex, columnName := range firstRow {
		if columnName == "" {
			return fmt.Errorf("autodetectColumns is enabled, but the column at index %v does not have a name.", columnIndex)
		}

		if _, alreadyExists := alreadyExistingNames[columnName]; alreadyExists {
			return fmt.Errorf("autodetectColumns is enabled, but there is a duplicate column named %v.", columnName)
		}
		alreadyExistingNames[columnName] = true

		csvInput.config.Columns[columnIndex] = &configModule.CsvInputColumn{
			Name:   columnName,
			Parser: "string",
		}
		csvInput.config.ColumnIndexByName[columnName] = columnIndex
	}

	return nil
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
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read csv position: %w", err)}
			}

			row, err := csvReader.Read()
			if err == io.EOF {
				break
			} else if errors.Is(err, csv.ErrFieldCount) {
				csvInput.config.Logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
			} else if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read csv data: %w", err)}
				return
			}

			channel <- IterateAllResult{
				Record: record.NewCsv(
					csvInput.config,
					csvInput.columnParsers,
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
