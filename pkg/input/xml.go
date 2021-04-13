package input

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	configModule "rodb.io/pkg/config"
	"rodb.io/pkg/parser"
	"rodb.io/pkg/record"
	"rodb.io/pkg/util"
	"sync"
)

var xmlParserOptions = xmlquery.ParserOptions{
	Decoder: &xmlquery.DecoderOptions{
		Strict: false,
	},
}

type Xml struct {
	config        *configModule.XmlInput
	reader        io.ReadSeeker
	readerBuffer  *bufio.Reader
	readerLock    sync.Mutex
	xmlFile       *os.File
	xmlParser     *xmlquery.StreamParser
	columnParsers []parser.Parser
	watcher       *fsnotify.Watcher
}

type xmlTempRecordNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr `xml:",any,attr"`
	InnerXML []byte     `xml:",innerxml"`
}

func NewXml(
	config *configModule.XmlInput,
	parsers parser.List,
) (*Xml, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	xmlInput := &Xml{
		config:     config,
		readerLock: sync.Mutex{},
		watcher:    watcher,
	}

	util.StartFilesystemWatchProcess(
		xmlInput.watcher,
		*xmlInput.config.DieOnInputChange,
		xmlInput.config.Logger,
	)

	file, err := os.Open(xmlInput.config.Path)
	if err != nil {
		return nil, err
	}

	xmlInput.xmlFile = file
	xmlInput.reader = io.ReadSeeker(file)
	xmlInput.readerBuffer = bufio.NewReader(xmlInput.reader)

	xmlInput.xmlParser, err = xmlquery.CreateStreamParserWithOptions(
		xmlInput.readerBuffer,
		xmlParserOptions,
		xmlInput.config.RecordXPath,
	)
	if err != nil {
		return nil, err
	}

	err = xmlInput.watcher.Add(config.Path)
	if err != nil {
		return nil, err
	}

	xmlInput.columnParsers = make([]parser.Parser, len(config.Columns))
	for i, column := range config.Columns {
		parser, parserExists := parsers[column.Parser]
		if !parserExists {
			return nil, errors.New("Parser '" + column.Parser + "' does not exist")
		}
		xmlInput.columnParsers[i] = parser
	}

	return xmlInput, nil
}

func (xmlInput *Xml) Name() string {
	return xmlInput.config.Name
}

func (xmlInput *Xml) HasColumn(columnName string) bool {
	for _, column := range xmlInput.config.Columns {
		if column.Name == columnName {
			return true
		}
	}

	return false
}

func (xmlInput *Xml) Get(position record.Position) (record.Record, error) {
	xmlInput.readerLock.Lock()
	defer xmlInput.readerLock.Unlock()

	util.SetBufferedReaderOffset(
		xmlInput.reader,
		xmlInput.readerBuffer,
		position,
	)

	node, err := xmlInput.xmlParser.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("Did not find an XML record at position %v", position)
	} else if err != nil {
		return nil, fmt.Errorf("Cannot read xml data: %w", err)
	}

	return record.NewXml(xmlInput.config, xmlInput.columnParsers, node, position)
}

func (xmlInput *Xml) Size() (int64, error) {
	fileInfo, err := os.Stat(xmlInput.config.Path)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (xmlInput *Xml) IterateAll() <-chan IterateAllResult {
	channel := make(chan IterateAllResult)

	go func() {
		defer close(channel)

		file, err := os.Open(xmlInput.config.Path)
		if err != nil {
			channel <- IterateAllResult{Error: err}
			return
		}
		defer file.Close()

		reader := io.ReadSeeker(file)
		readerBuffer := bufio.NewReader(reader)

		xmlParser, err := xmlquery.CreateStreamParserWithOptions(
			readerBuffer,
			xmlParserOptions,
			xmlInput.config.RecordXPath,
		)
		if err != nil {
			channel <- IterateAllResult{Error: err}
			return
		}

		position := int64(0)
		for {
			// The returned position is actually the end of the opening tag of the previous element
			// We cannot easily get a more precise position, but since we can get the next record
			// reliably from this position, it works
			position, err = util.GetBufferedReaderOffset(reader, readerBuffer)
			if err != nil {
				channel <- IterateAllResult{
					Error: fmt.Errorf("Error when getting xml offset: %v", err),
				}
				return
			}

			node, err := xmlParser.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read xml data: %w", err)}
				return
			}

			record, err := record.NewXml(xmlInput.config, xmlInput.columnParsers, node, position)
			if err != nil {
				channel <- IterateAllResult{
					Error: fmt.Errorf("Error when creating record after position %v: %v", position, err),
				}
				return
			}

			channel <- IterateAllResult{Record: record}
		}
	}()

	return channel
}

func (xmlInput *Xml) Close() error {
	err := xmlInput.watcher.Remove(xmlInput.config.Path)
	if err != nil {
		return err
	}

	err = xmlInput.watcher.Close()
	if err != nil {
		return err
	}

	err = xmlInput.xmlFile.Close()
	if err != nil {
		return err
	}

	return nil
}
