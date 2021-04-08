package input

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	configModule "rods/pkg/config"
	"rods/pkg/parser"
	"rods/pkg/record"
	"rods/pkg/util"
	"sync"
)

type Csv struct {
	config          *configModule.CsvInput
	reader          io.ReadSeeker
	readerLock      sync.Mutex
	csvFile         *os.File
	csvReader       *csv.Reader
	csvReaderBuffer *bufio.Reader
	columnParsers   []parser.Parser
	watcher         *fsnotify.Watcher
}

func NewCsv(
	config *configModule.CsvInput,
	parsers parser.List,
) (*Csv, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	csvInput := &Csv{
		config:     config,
		readerLock: sync.Mutex{},
		watcher:    watcher,
	}

	util.StartFilesystemWatchProcess(
		csvInput.watcher,
		*csvInput.config.DieOnInputChange,
		csvInput.config.Logger,
	)

	reader, csvReader, file, err := csvInput.open()
	if err != nil {
		return nil, err
	}
	csvInput.reader = reader
	csvInput.csvFile = file
	csvInput.csvReader = csvReader
	csvInput.csvReaderBuffer = util.GetInternalBufferReader(csvReader, "r")

	err = csvInput.watcher.Add(config.Path)
	if err != nil {
		return nil, err
	}

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
	csvInput.readerLock.Lock()
	defer csvInput.readerLock.Unlock()

	util.SetBufferedReaderOffset(
		csvInput.reader,
		csvInput.csvReaderBuffer,
		position,
	)

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
	fileInfo, err := os.Stat(csvInput.config.Path)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
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

func (csvInput *Csv) open() (io.ReadSeeker, *csv.Reader, *os.File, error) {
	file, err := os.Open(csvInput.config.Path)
	if err != nil {
		return nil, nil, nil, err
	}

	reader := io.ReadSeeker(file)

	csvReader := csv.NewReader(reader)
	csvReader.Comma = []rune(csvInput.config.Delimiter)[0]
	csvReader.FieldsPerRecord = len(csvInput.config.Columns)
	csvReader.ReuseRecord = false

	return reader, csvReader, file, nil
}

func (csvInput *Csv) IterateAll() <-chan IterateAllResult {
	channel := make(chan IterateAllResult)

	go func() {
		defer close(channel)

		reader, csvReader, file, err := csvInput.open()
		if err != nil {
			channel <- IterateAllResult{Error: err}
			return
		}
		defer file.Close()

		if csvInput.config.IgnoreFirstRow {
			_, _ = csvReader.Read()
		}

		csvReaderBuffer := util.GetInternalBufferReader(csvReader, "r")

		for {
			position, err := util.GetBufferedReaderOffset(
				reader,
				csvReaderBuffer,
			)
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
	err := csvInput.watcher.Remove(csvInput.config.Path)
	if err != nil {
		return err
	}

	err = csvInput.watcher.Close()
	if err != nil {
		return err
	}

	err = csvInput.csvFile.Close()
	if err != nil {
		return err
	}

	return nil
}
