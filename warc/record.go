package warc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// Record is a WARC record.
type Record struct {
	// Header is a WARC header containing information about the payload. Guaranteed to exist.
	Header *Header
	// Payload is the body of the record. May not exist.
	Payload *Payload
}

// ReadRecord reads a record. Returns nil for the record if none was read (EOF).
func ReadRecord(reader *bufio.Reader) (*Record, error) {
	header, err := ReadHeader(reader)
	if err != nil {
		return nil, err
	}

	payload, err := ReadPayload(reader, header)
	if err != nil {
		return nil, err
	}

	record := &Record{
		Header:  header,
		Payload: payload,
	}

	return record, nil
}

// ReadRecordHeader works like ReadRecord, but always skips the payload.
func ReadRecordHeader(reader *bufio.Reader) (*Record, error) {
	header, err := ReadHeader(reader)
	if err != nil {
		return nil, err
	}

	if header.ContentLength > 0 {
		// TODO: Add support for uint64 by looping?
		// Skip the 2x CRLF after the payload (+4)
		discarded, err := reader.Discard(int(header.ContentLength) + 4)
		if err != nil {
			return nil, fmt.Errorf("Unable to discard CRLF after header: %v", err)
		}
		if discarded != int(header.ContentLength)+4 {
			return nil, fmt.Errorf("Expected to skip %vB, but skipped %vB", header.ContentLength, discarded)
		}
	}

	record := &Record{
		Header:  header,
		Payload: nil,
	}

	return record, nil
}

// Write writes the record to a stream.
func (record *Record) Write(writer io.Writer) error {
	err := record.Header.Write(writer)
	if err != nil {
		return err
	}

	writer.Write([]byte("\r\n"))

	if record.Payload != nil {
		record.Payload.Write(writer)
	}

	return nil
}

// String converts the record into a string
func (record *Record) String() (string, error) {
	buffer := new(bytes.Buffer)
	err := record.Write(buffer)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
