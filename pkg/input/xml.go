package input

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/antchfx/xmlquery"
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

var xmlParserOptions = xmlquery.ParserOptions{
	Decoder: &xmlquery.DecoderOptions{
		Strict: false,
	},
}

type Xml struct {
	config       *configPackage.XmlInput
	reader       io.ReadSeeker
	readerBuffer *bufio.Reader
	readerLock   sync.Mutex
	xmlFile      *os.File
	xmlParser    *xmlquery.StreamParser
	parsers      parser.List
	watcher      *fsnotify.Watcher
}

type xmlTempRecordNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr `xml:",any,attr"`
	InnerXML []byte     `xml:",innerxml"`
}

func NewXml(
	config *configPackage.XmlInput,
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
		parsers:    parsers,
	}

	util.StartFilesystemWatchProcess(
		xmlInput.watcher,
		xmlInput.config.ShouldDieOnInputChange(),
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

	if err := xmlInput.watcher.Add(config.Path); err != nil {
		return nil, err
	}

	return xmlInput, nil
}

func (xmlInput *Xml) Name() string {
	return xmlInput.config.Name
}

func (xmlInput *Xml) Get(position record.Position) (record.Record, error) {
	xmlInput.readerLock.Lock()
	defer xmlInput.readerLock.Unlock()

	if err := util.SetBufferedReaderOffset(xmlInput.reader, xmlInput.readerBuffer, position); err != nil {
		return nil, err
	}

	node, err := xmlInput.xmlParser.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("Did not find an XML record at position %v", position)
	} else if err != nil {
		return nil, fmt.Errorf("Cannot read xml data: %w", err)
	}

	return record.NewXml(xmlInput.config, node, xmlInput.parsers, position)
}

func (xmlInput *Xml) Size() (int64, error) {
	fileInfo, err := os.Stat(xmlInput.config.Path)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (xmlInput *Xml) ModTime() (time.Time, error) {
	fileInfo, err := os.Stat(xmlInput.config.Path)
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime(), nil
}

func (xmlInput *Xml) IterateAll() (record.Iterator, func() error, error) {
	file, err := os.Open(xmlInput.config.Path)
	if err != nil {
		return nil, nil, err
	}

	reader := io.ReadSeeker(file)
	readerBuffer := bufio.NewReader(reader)

	xmlParser, err := xmlquery.CreateStreamParserWithOptions(
		readerBuffer,
		xmlParserOptions,
		xmlInput.config.RecordXPath,
	)
	if err != nil {
		return nil, nil, err
	}

	position := int64(0)
	iterator := func() (record.Record, error) {
		// The returned position is actually the end of the opening tag of the previous element
		// We cannot easily get a more precise position, but since we can get the next record
		// reliably from this position, it works
		position, err = util.GetBufferedReaderOffset(reader, readerBuffer)
		if err != nil {
			return nil, fmt.Errorf("Error when getting xml offset: %v", err)
		}

		node, err := xmlParser.Read()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("Cannot read xml data: %w", err)
		}

		record, err := record.NewXml(xmlInput.config, node, xmlInput.parsers, position)
		if err != nil {
			return nil, fmt.Errorf("Error when creating record after position %v: %v", position, err)
		}

		return record, nil
	}

	end := func() error {
		return file.Close()
	}

	return iterator, end, nil
}

func (xmlInput *Xml) Close() error {
	if err := xmlInput.watcher.Remove(xmlInput.config.Path); err != nil {
		return err
	}

	if err := xmlInput.watcher.Close(); err != nil {
		return err
	}

	if err := xmlInput.xmlFile.Close(); err != nil {
		return err
	}

	return nil
}
