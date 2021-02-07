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
// TODO: give the payload parser a "scoped" reader, with bounds only for the payload?
type PayloadParser func(reader *bufio.Reader, header *Header) (IPayload, error)

// Reader is a WARC reader.
type Reader struct {
	stringReader   *bufio.Reader
	reader         io.ReadSeeker
	payloadParsers map[string]PayloadParser
	Seekable       bool
}

// NewReader creates a new reader. Compressed specifies whether or not
// the stream is compressed using gzip.
func NewReader(reader io.ReadSeeker, compressed bool) (*Reader, error) {
	var stringReader *bufio.Reader
	if compressed {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		stringReader = bufio.NewReader(gzipReader)
	} else {
		stringReader = bufio.NewReader(reader)
	}

	fileReader := &Reader{
		reader:         reader,
		stringReader:   stringReader,
		payloadParsers: make(map[string]PayloadParser),
		Seekable:       !compressed,
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
	buffer, _, err := reader.stringReader.ReadLine()
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Unable to read line for header: %v", err)
	}

	line := string(buffer)
	if line != "WARC/1.0" {
		return nil, fmt.Errorf("Expected WARC version header 'WARC/1.0' got '%v'", line)
	}

	err = UnmarshalStream(reader.stringReader, header)
	if err != nil {
		return nil, err
	}

	// Since we're using a buffered reader, it will have read more than
	// the file reader knows - meaning, the tell here will be incorrect...
	// Store the offset to the header's payload in the archive file
	payloadOffset, err := reader.reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	header.payloadOffset = payloadOffset - int64(reader.stringReader.Buffered())

	return header, nil
}

// ReadPayload reads a payload. Returns nil if none was read (EOF).
// May be invoked at any time as the function seeks to the correct
// place in the archive to read the payload.
func (reader *Reader) ReadPayload(header *Header) (IPayload, error) {
	// No payload to read
	if header.ContentLength <= 0 {
		return nil, nil
	}

	var offset int64
	if reader.Seekable {
		// Get the current offset in the file
		// Note that offset is not restored if there's an error in this function
		_offset, err := reader.reader.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}

		// Seek to the start of the payload
		_, err = reader.reader.Seek(header.payloadOffset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		reader.stringReader.Reset(reader.reader)

		offset = _offset
	}

	// Get the registered parser, defaulting to a raw payload
	parser := reader.payloadParsers[header.Type]
	if parser == nil {
		parser = reader.payloadParsers[AnyRecordType]
	}

	if parser == nil {
		return nil, fmt.Errorf("No parser registered for record type '%s'", header.Type)
	}

	payload, err := parser(reader.stringReader, header)
	if err != nil {
		return nil, err
	}

	// Read 2x CRLF after the payload
	_, err = reader.stringReader.Discard(4)
	// Ignoring EOF for more permissive parsing
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("Unable to discard CRLF after payload: %v", err)
	}

	if reader.Seekable {
		// Seek to the previous offset
		_, err = reader.reader.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		reader.stringReader.Reset(reader.reader)
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
		_, err := reader.stringReader.Discard(int(header.ContentLength) + 4)
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

// ReadAllHeaders reads all records' headers, excluding their payloads.
func (reader *Reader) ReadAllHeaders() (*File, error) {
	records := make([]*Record, 0)

	for {
		record, err := reader.ReadRecordHeader()
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
