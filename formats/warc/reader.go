package warc

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
)

// AnyRecordType depicts a catch all record type.
const AnyRecordType = "*"

// PayloadParser is a parser invoked to unmarshal a payload.
type PayloadParser func(reader *bufio.Reader, header *Header) (IPayload, error)

// Reader is a WARC reader.
type Reader struct {
	reader         *bufio.Reader
	payloadParsers map[string]PayloadParser
}

// NewReader creates a new reader. Compressed specifies whether or not
// the stream is compressed using gzip.
func NewReader(reader io.Reader, compressed bool) (*Reader, error) {
	if compressed {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		reader = bufio.NewReader(gzipReader)
	}

	fileReader := &Reader{
		reader:         bufio.NewReader(reader),
		payloadParsers: make(map[string]PayloadParser),
	}

	// Set a default payload parser
	fileReader.RegisterPayloadParser(AnyRecordType, RawPayloadParser)

	return fileReader, nil
}

// RegisterPayloadParser registers a parser to be invoked after reading a header
// for the specific recordType.
// An asterix ('*') may be used to register the any parser. By default this
// parser will parse payloads as RawPayload.
func (reader *Reader) RegisterPayloadParser(recordType string, parser PayloadParser) {
	reader.payloadParsers[recordType] = parser
}

// ReadHeader reads a header. Returns nil if none was read (EOF).
func (reader *Reader) ReadHeader() (*Header, error) {
	header := &Header{}

	// Read the version - WARC/1.0
	buffer, _, err := reader.reader.ReadLine()
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Unable to read line for header: %v", err)
	}

	line := string(buffer)
	if line != "WARC/1.0" {
		return nil, fmt.Errorf("Expected WARC version header 'WARC/1.0' got '%v'", line)
	}

	err = UnmarshalStream(reader.reader, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// ReadPayload reads a payload. Returns nil if none was read (EOF).
func (reader *Reader) ReadPayload(header *Header) (IPayload, error) {
	if header.ContentLength <= 0 {
		return nil, nil
	}

	// Get the registered parser, defaulting to a raw payload
	parser := reader.payloadParsers[header.Type]
	if parser == nil {
		parser = reader.payloadParsers[AnyRecordType]
	}

	if parser == nil {
		return nil, fmt.Errorf("No parser registered for record type '%s'", header.Type)
	}

	payload, err := parser(reader.reader, header)
	if err != nil {
		return nil, err
	}

	// Read 2x CRLF after the payload
	_, err = reader.reader.Discard(4)
	// Ignoring EOF for more permissive parsing
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("Unable to discard CRLF after payload: %v", err)
	}

	return payload, nil
}

// ReadRecordHeader reads the next record, skipping the payload.
func (reader *Reader) ReadRecordHeader() (*Record, error) {
	header, err := reader.ReadHeader()
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, nil
	}

	if header.ContentLength > 0 {
		// TODO: Add support for uint64 by looping?
		// Skip the 2x CRLF after the payload (+4)
		_, err := reader.reader.Discard(int(header.ContentLength) + 4)
		// Ignoring EOF for more permissive parsing
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("Unable to discard CRLF after header: %v", err)
		}
	}

	record := &Record{
		Header:  header,
		Payload: nil,
	}

	return record, nil
}

// ReadRecord reads the next record.
func (reader *Reader) ReadRecord() (*Record, error) {
	header, err := reader.ReadHeader()
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, nil
	}

	record := &Record{
		Header:  header,
		Payload: nil,
	}

	if header.ContentLength > 0 {
		payload, err := reader.ReadPayload(header)
		if err != nil {
			return nil, err
		}
		record.Payload = payload
	}

	return record, nil
}

// ReadAll reads all records, including their payloads.
func (reader *Reader) ReadAll() (*File, error) {
	records := make([]*Record, 0)

	for {
		record, err := reader.ReadRecord()
		if err != nil {
			return nil, err
		}

		if record == nil {
			break
		}

		records = append(records, record)
	}

	return &File{
		Records: records,
	}, nil
}
