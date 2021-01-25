package warc

import (
	"fmt"
	"io"
	"time"
)

// Payload is any payload of a WARC Record
type Payload struct {
	// Data is the raw data of the payload
	Data []byte
	// Length is the length of the data in bytes
	Length uint64
}

// InfoPayload is the payload of a "warcinfo" record.
type InfoPayload struct {
	Payload
	// Operator contains contact information for the operator who created this WARC resource.
	Operator string `warc:"operator"`
	// Software is the software and version used to create this WARC resource.
	Software string `warc:"software"`
	// Robots is the robots policy followed by the harvester creating this WARC resource.
	Robots string `warc:"robots"`
	// Hostname is the hostname of the machine that created this WARC resource.
	Hostname string `warc:"hostname"`
	// The IP address of the machine that created this WARC resource.
	IP string `warc:"ip"`
	// UserAgent is the HTTP 'user-agent' header usually sent by the harvester along with each request.
	UserAgent string `warc:"http-header-user-agent"`
	// From is the HTTP 'From' header usually sent by the harvester along with each request.
	From string `warc:"http-header-from"`
}

// MetadataPayload is a payload record that contains content created in order to further describe a harvested resource.
type MetadataPayload struct {
	Payload
	// Via is the referring URI from which the archived URI was discorvered.
	Via string `warc:"via"`
	// HopsFromSeed describes the type of each hop from a starting URI to the current URI.
	HopsFromSeed string `warc:"hopsFromSeed"`
	// FetchTime is the time that it took to collect the archived URI, starting from the initation of network traffic.
	FetchTime time.Duration `warc:"fetchTimeMs"`
}

// ReadPayload reads the payload of a record.
func ReadPayload(reader io.ReadSeeker, header *Header) (*Payload, error) {
	if header.ContentLength <= 0 {
		return nil, nil
	}

	payload := &Payload{
		Data:   make([]byte, header.ContentLength),
		Length: header.ContentLength,
	}

	bytesRead, err := io.ReadFull(reader, payload.Data)
	if err != nil {
		return nil, err
	}

	if bytesRead != int(header.ContentLength) {
		return nil, fmt.Errorf("Unable to read payload. Expected %v bytes got %v", header.ContentLength, bytesRead)
	}

	payload, err = ParsePayload(payload, header)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// ParsePayload parses a single payload if it's of a supported type. Leaves it unchanged otherwise.
func ParsePayload(payload *Payload, header *Header) (*Payload, error) {
	if header.Type == TypeInfo {
		return ParseInfoPayload(payload, header)
	} else if header.Type == TypeMetadata {
		return ParseMetadataPayload(payload, header)
	}

	return payload, nil
}

// ParseInfoPayload parses a WARC info record's payload.
func ParseInfoPayload(payload *Payload, header *Header) (*Payload, error) {
	// TODO: Actually parse payload
	return payload, nil
}

// ParseInfoPayload parses a WARC metadata record's payload.
func ParseMetadataPayload(payload *Payload, header *Header) (*Payload, error) {
	// TODO: Actually parse payload
	return payload, nil
}