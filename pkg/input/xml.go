package input

import (
	"bufio"
	"encoding/xml"
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

type Xml struct {
	config           *configModule.XmlInput
	reader           io.ReadSeeker
	readerLock       sync.Mutex
	xmlFile          *os.File
	xmlDecoder       *xml.Decoder
	xmlDecoderBuffer *bufio.Reader
	columnParsers    []parser.Parser
	watcher          *fsnotify.Watcher
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
	xmlInput.xmlDecoder = xml.NewDecoder(xmlInput.reader)
	xmlInput.xmlDecoderBuffer = util.GetInternalBufferReader(xmlInput.xmlDecoder, "r")

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
		xmlInput.xmlDecoderBuffer,
		position,
	)

	token, err := xmlInput.xmlDecoder.Token()
	if token == nil || err == io.EOF {
		return nil, fmt.Errorf("Did not find an XML record at position %v", position)
	} else if err != nil {
		return nil, fmt.Errorf("Cannot read xml data: %w", err)
	}

	element, isStartElement := token.(xml.StartElement)
	if !isStartElement {
		return nil, fmt.Errorf("Did not find an XML opening tag at position %v", position)
	}

	if element.Name.Local != xmlInput.config.ElementNodeName {
		return nil, fmt.Errorf("The tag at position %v is not a '%v'.", position, xmlInput.config.ElementNodeName)
	}

	xmlData, err := xmlInput.getOuterXml(xmlInput.xmlDecoder, element)
	if err != nil {
		return nil, err
	}

	return record.NewXml(xmlInput.config, xmlInput.columnParsers, xmlData, position), nil
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
		xmlDecoder := xml.NewDecoder(reader)
		xmlDecoderBuffer := util.GetInternalBufferReader(xmlDecoder, "r")

		position := int64(0)
		for {
			token, err := xmlDecoder.Token()
			if token == nil || err == io.EOF {
				break
			} else if err != nil {
				channel <- IterateAllResult{Error: fmt.Errorf("Cannot read xml data: %w", err)}
				return
			}

			element, isStartElement := token.(xml.StartElement)
			if isStartElement {
				if element.Name.Local == xmlInput.config.ElementNodeName {
					result, err := xmlInput.getOuterXml(xmlDecoder, element)
					if err != nil {
						channel <- IterateAllResult{
							Error: fmt.Errorf("Error when reading xml record after position %v: %v", position, err),
						}
						return
					}

					channel <- IterateAllResult{
						Record: record.NewXml(xmlInput.config, xmlInput.columnParsers, result, position),
					}
				}
			}

			// The position must be updated here, because the one that allows
			// retrieving an item in one operation is the one after the end
			// of the previous item.
			position, err = util.GetBufferedReaderOffset(
				reader,
				xmlDecoderBuffer,
			)
			if err != nil {
				channel <- IterateAllResult{
					Error: fmt.Errorf("Error when getting xml offset: %v", err),
				}
				return
			}

			if position > 0 {
				// The cursor must be placed one byte earlier
				// any future read operation to work as expected.
				position--
			}
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

func (xmlInput *Xml) getOuterXml(
	decoder *xml.Decoder,
	element xml.StartElement,
) ([]byte, error) {
	result := xmlTempRecordNode{}
	err := decoder.DecodeElement(&result, &element)
	if err != nil {
		return nil, err
	}

	outerXml, err := xml.Marshal(result)
	if err != nil {
		return nil, err
	}

	return outerXml, nil
}
