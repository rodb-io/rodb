package input

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	configPackage "rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
	"sync"
	"time"
)

type Csv struct {
	config        *configPackage.CsvInput
	reader        io.ReadSeeker
	readerLock    sync.Mutex
	csvFile       *os.File
	csvReader     *csv.Reader
	readerBuffer  *bufio.Reader
	columnParsers []parser.Parser
	watcher       *fsnotify.Watcher
}

func NewCsv(
	config *configPackage.CsvInput,
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
		csvInput.config.ShouldDieOnInputChange(),
		csvInput.config.Logger,
	)

	reader, readerBuffer, csvReader, file, err := csvInput.open()
	if err != nil {
		return nil, err
	}
	csvInput.reader = reader
	csvInput.readerBuffer = readerBuffer
	csvInput.csvFile = file
	csvInput.csvReader = csvReader

	if err := csvInput.watcher.Add(config.Path); err != nil {
		return nil, err
	}

	if config.AutodetectColumns {
		if err := csvInput.autodetectColumns(); err != nil {
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

func (csvInput *Csv) Get(position record.Position) (record.Record, error) {
	csvInput.readerLock.Lock()
	defer csvInput.readerLock.Unlock()

	if err := util.SetBufferedReaderOffset(csvInput.reader, csvInput.readerBuffer, position); err != nil {
		return nil, err
	}

	row, err := csvInput.csvReader.Read()
	if err != nil {
		if errors.Is(err, csv.ErrFieldCount) {
			csvInput.config.Logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
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

func (csvInput *Csv) ModTime() (time.Time, error) {
	fileInfo, err := os.Stat(csvInput.config.Path)
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime(), nil
}

func (csvInput *Csv) autodetectColumns() error {
	firstRow, err := csvInput.csvReader.Read()
	if err != nil {
		return fmt.Errorf("Cannot read csv data: %w", err)
	}

	alreadyExistingNames := make(map[string]bool)
	csvInput.config.Columns = make([]*configPackage.CsvInputColumn, len(firstRow))
	csvInput.config.ColumnIndexByName = make(map[string]int)
	for columnIndex, columnName := range firstRow {
		if columnName == "" {
			return fmt.Errorf("autodetectColumns is enabled, but the column at index %v does not have a name.", columnIndex)
		}

		if _, alreadyExists := alreadyExistingNames[columnName]; alreadyExists {
			return fmt.Errorf("autodetectColumns is enabled, but there is a duplicate column named %v.", columnName)
		}
		alreadyExistingNames[columnName] = true

		csvInput.config.Columns[columnIndex] = &configPackage.CsvInputColumn{
			Name:   columnName,
			Parser: "string",
		}
		csvInput.config.ColumnIndexByName[columnName] = columnIndex
	}

	return nil
}

func (csvInput *Csv) open() (io.ReadSeeker, *bufio.Reader, *csv.Reader, *os.File, error) {
	file, err := os.Open(csvInput.config.Path)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	reader := io.ReadSeeker(file)

	// Giving a buffer to the csv reader will prevent it from creating
	// it's own buffer, since we need to control it when seeking
	// the positions (this condition is managed by bufio's constructor)
	readerBuffer := bufio.NewReader(reader)

	csvReader := csv.NewReader(readerBuffer)
	csvReader.Comma = []rune(csvInput.config.Delimiter)[0]
	csvReader.FieldsPerRecord = len(csvInput.config.Columns)
	csvReader.ReuseRecord = false

	return reader, readerBuffer, csvReader, file, nil
}

func (csvInput *Csv) IterateAll() (record.Iterator, func() error, error) {
	reader, readerBuffer, csvReader, file, err := csvInput.open()
	if err != nil {
		return nil, nil, err
	}

	if csvInput.config.IgnoreFirstRow {
		_, err = csvReader.Read()
		if err != nil {
			return nil, nil, err
		}
	}

	iterator := func() (record.Record, error) {
		position, err := util.GetBufferedReaderOffset(reader, readerBuffer)
		if err != nil {
			return nil, fmt.Errorf("Cannot read csv position: %w", err)
		}

		row, err := csvReader.Read()
		if err == io.EOF {
			return nil, nil
		} else if errors.Is(err, csv.ErrFieldCount) {
			csvInput.config.Logger.Warnf("Expected %v columns in csv, got %+v", len(csvInput.config.Columns), row)
		} else if err != nil {
			return nil, fmt.Errorf("Cannot read csv data: %w", err)
		}

		record := record.NewCsv(
			csvInput.config,
			csvInput.columnParsers,
			row,
			position,
		)

		return record, nil
	}

	end := func() error {
		return file.Close()
	}

	return iterator, end, nil
}

func (csvInput *Csv) Close() error {
	if err := csvInput.watcher.Remove(csvInput.config.Path); err != nil {
		return err
	}

	if err := csvInput.watcher.Close(); err != nil {
		return err
	}

	if err := csvInput.csvFile.Close(); err != nil {
		return err
	}

	return nil
}
