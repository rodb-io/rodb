package input

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	configModule "rodb.io/pkg/config"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
	"sync"
)

type Json struct {
	config     *configModule.JsonInput
	reader     io.ReadSeeker
	readerLock sync.Mutex
	jsonFile   *os.File
	watcher    *fsnotify.Watcher
}

func NewJson(config *configModule.JsonInput) (*Json, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	jsonInput := &Json{
		config:     config,
		readerLock: sync.Mutex{},
		watcher:    watcher,
	}

	util.StartFilesystemWatchProcess(
		jsonInput.watcher,
		*jsonInput.config.DieOnInputChange,
		jsonInput.config.Logger,
	)

	reader, file, err := jsonInput.open()
	if err != nil {
		return nil, err
	}
	jsonInput.reader = reader
	jsonInput.jsonFile = file

	// Returning to the beginning of the file after checking the first token
	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	err = jsonInput.watcher.Add(config.Path)
	if err != nil {
		return nil, err
	}

	return jsonInput, nil
}

func (jsonInput *Json) Name() string {
	return jsonInput.config.Name
}

func (jsonInput *Json) Get(position record.Position) (record.Record, error) {
	jsonInput.readerLock.Lock()
	defer jsonInput.readerLock.Unlock()

	_, err := jsonInput.reader.Seek(position, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("Cannot seek position '%v' in json file: %w", position, err)
	}

	// Creating a new decoder each time because we cannot reset
	// it's internal state when seeking otherwise
	jsonDecoder := json.NewDecoder(jsonInput.reader)

	var data map[string]interface{}
	err = jsonDecoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("Cannot read json data: %w", err)
	}

	return record.NewJson(
		jsonInput.config,
		data,
		position,
	), nil
}

func (jsonInput *Json) Size() (int64, error) {
	fileInfo, err := os.Stat(jsonInput.config.Path)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (jsonInput *Json) open() (io.ReadSeeker, *os.File, error) {
	file, err := os.Open(jsonInput.config.Path)
	if err != nil {
		return nil, nil, err
	}

	reader := io.ReadSeeker(file)

	return reader, file, nil
}

func (jsonInput *Json) IterateAll() (record.Iterator, func() error, error) {
	reader, file, err := jsonInput.open()
	if err != nil {
		return nil, nil, err
	}

	// Creating a new decoder each time because we cannot reset
	// it's internal state when seeking otherwise
	jsonDecoder := json.NewDecoder(reader)

	iterator := func() (record.Record, error) {
		if !jsonDecoder.More() {
			return nil, nil
		}

		position := jsonDecoder.InputOffset()

		var data map[string]interface{}
		err = jsonDecoder.Decode(&data)
		if err == io.EOF {
			return nil, nil
		} else if err != nil {
			return nil, fmt.Errorf("Cannot read json data: %w", err)
		}

		record := record.NewJson(
			jsonInput.config,
			data,
			position,
		)

		return record, nil
	}

	end := func() error {
		return file.Close()
	}

	return iterator, end, nil
}

func (jsonInput *Json) Close() error {
	err := jsonInput.watcher.Remove(jsonInput.config.Path)
	if err != nil {
		return err
	}

	err = jsonInput.watcher.Close()
	if err != nil {
		return err
	}

	err = jsonInput.jsonFile.Close()
	if err != nil {
		return err
	}

	return nil
}
