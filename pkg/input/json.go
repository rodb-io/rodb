package input

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	configModule "rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
	"sync"
)

type Json struct {
	config      *configModule.JsonInput
	reader      io.ReadSeeker
	readerLock  sync.Mutex
	jsonFile    *os.File
	jsonDecoder *json.Decoder
	watcher     *fsnotify.Watcher

	// If true, the file is a json array containing the required objects
	// If false, the file contains multiple independent json documents
	jsonIsArray bool
}

func NewJson(
	config *configModule.JsonInput,
	parsers parser.List,
) (*Json, error) {
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

	reader, jsonDecoder, file, err := jsonInput.open()
	if err != nil {
		return nil, err
	}
	jsonInput.reader = reader
	jsonInput.jsonFile = file
	jsonInput.jsonDecoder = jsonDecoder

	// Analyzing the first token to check if it's an array
	// or a sequence of documents
	firstToken, err := jsonDecoder.Token()
	if err != nil {
		return nil, err
	}
	jsonInput.jsonIsArray = (firstToken == json.Delim('['))

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

	var data map[string]interface{}
	err = jsonInput.jsonDecoder.Decode(&data)
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

func (jsonInput *Json) open() (io.ReadSeeker, *json.Decoder, *os.File, error) {
	file, err := os.Open(jsonInput.config.Path)
	if err != nil {
		return nil, nil, nil, err
	}

	reader := io.ReadSeeker(file)

	jsonDecoder := json.NewDecoder(reader)

	return reader, jsonDecoder, file, nil
}

func (jsonInput *Json) IterateAll() (record.Iterator, func() error, error) {
	reader, jsonDecoder, file, err := jsonInput.open()
	if err != nil {
		return nil, nil, err
	}

	if jsonInput.jsonIsArray {
		// Skipping the first token (array opening)
		_, err := jsonInput.jsonDecoder.Token()
		if err != nil {
			return nil, nil, err
		}
	}

	iterator := func() (record.Record, error) {
		position, err := reader.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("Cannot read json position: %w", err)
		}

		if !jsonDecoder.More() {
			return nil, nil
		}

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
