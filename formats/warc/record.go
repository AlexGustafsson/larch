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
	Payload IPayload
}

// Write writes the record to a stream.
func (record *Record) Write(writer io.Writer) error {
	err := record.Header.Write(writer)
	if err != nil {
		return err
	}

	writer.Write([]byte("\r\n"))

	if record.Payload != nil {
		err = record.Payload.Write(writer)
		if err != nil {
			return err
		}
	}

	writer.Write([]byte("\r\n\r\n"))

	return nil
}

// WriteHeader writes the record's header and an empty payload to a stream.
func (record *Record) WriteHeader(writer io.Writer) error {
	// Create a copy of the record header, zeroing the content length
	// TODO: Investigate less memory-intensive solutions
	dereferencedHeader := *record.Header
	headerCopy := dereferencedHeader
	headerCopy.ContentLength = 0

	err := headerCopy.Write(writer)
	if err != nil {
		return err
	}

	writer.Write([]byte("\r\n\r\n\r\n"))

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
