package warc

import (
	"bytes"
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
func ReadRecord(reader io.ReadSeeker) (*Record, error) {
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

// Write writes the record to a stream.
func (record *Record) Write(writer io.Writer) {
	record.Header.Write(writer)
	record.Payload.Write(writer)
}

// String converts the record into a string
func (record *Record) String() string {
	buffer := new(bytes.Buffer)
	record.Write(buffer)
	return buffer.String()
}