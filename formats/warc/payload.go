package warc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"
)

// IPayload is any payload of a WARC Record.
type IPayload interface {
	// Write writes the payload to a stream.
	Write(writer io.Writer) error
	// Bytes returns the byte representation of the payload.
	Bytes() ([]byte, error)
	// String converts the payload into a string.
	String() (string, error)
	// Reader returns a reader for the data.
	Reader() io.Reader
}

// RawPayload is a base IPayload implementation for raw bytes.
type RawPayload struct {
	// Data is the raw data of the payload
	Data []byte
	// Length is the length of the data in bytes
	Length uint64
}

// Write writes the payload to a stream.
func (payload *RawPayload) Write(writer io.Writer) error {
	bytesWritten, err := writer.Write(payload.Data)
	if err != nil {
		return err
	}

	if uint64(bytesWritten) != payload.Length {
		return fmt.Errorf("Expected to write %d bytes, wrote %d", payload.Length, bytesWritten)
	}

	return nil
}

// Bytes returns the byte representation of the payload.
func (payload *RawPayload) Bytes() ([]byte, error) {
	return payload.Data, nil
}

// String converts the payload into a string.
func (payload *RawPayload) String() (string, error) {
	buffer := new(bytes.Buffer)
	err := payload.Write(buffer)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// Reader returns a reader for the data.
func (payload *RawPayload) Reader() io.Reader {
	return bytes.NewReader(payload.Data)
}

// InfoPayload is the payload of a "warcinfo" record.
type InfoPayload struct {
	RawPayload
	// Operator contains contact information for the operator who created this WARC resource.
	Operator string `warc:"operator,omitempty"`
	// Software is the software and version used to create this WARC resource.
	Software string `warc:"software,omitempty"`
	// Robots is the robots policy followed by the harvester creating this WARC resource.
	Robots string `warc:"robots,omitempty"`
	// Hostname is the hostname of the machine that created this WARC resource.
	Hostname string `warc:"hostname,omitempty"`
	// The IP address of the machine that created this WARC resource.
	IP string `warc:"ip,omitempty"`
	// UserAgent is the HTTP 'user-agent' header usually sent by the harvester along with each request.
	UserAgent string `warc:"http-header-user-agent,omitempty"`
	// From is the HTTP 'From' header usually sent by the harvester along with each request.
	From string `warc:"http-header-from,omitempty"`
}

// MetadataPayload is a payload record that contains content created in order to further describe a harvested resource.
type MetadataPayload struct {
	RawPayload
	// Via is the referring URI from which the archived URI was discorvered.
	Via string `warc:"via,omitempty"`
	// HopsFromSeed describes the type of each hop from a starting URI to the current URI.
	HopsFromSeed string `warc:"hopsFromSeed,omitempty"`
	// FetchTime is the time that it took to collect the archived URI, starting from the initation of network traffic.
	FetchTime time.Duration `warc:"fetchTimeMs"`
}

// RawPayloadParser is a payload parser for reading raw payloads.
func RawPayloadParser(reader *bufio.Reader, header *Header) (IPayload, error) {
	payload := &RawPayload{
		Data:   nil,
		Length: header.ContentLength,
	}

	buffer := make([]byte, header.ContentLength)
	_, err := io.ReadFull(reader, buffer)
	if err != nil {
		return nil, fmt.Errorf("Unable to read payload buffer: %v", err)
	}

	payload.Data = buffer

	return payload, nil
}

// ParseInfoPayload parses a WARC info record's payload.
func ParseInfoPayload(payload IPayload, header *Header) (IPayload, error) {
	// TODO: Actually parse payload
	return payload, nil
}

// ParseMetadataPayload parses a WARC metadata record's payload.
func ParseMetadataPayload(payload IPayload, header *Header) (IPayload, error) {
	// TODO: Actually parse payload
	return payload, nil
}
