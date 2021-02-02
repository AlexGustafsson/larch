package payloads

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// Metadata is a Larch-specific record payload for metadata about the archival process.
type Metadata struct {
	Targets []string
}

// NewMetadata creates a new Metadata payload.
func NewMetadata() *Metadata {
	return &Metadata{
		Targets: make([]string, 0),
	}
}

// ReadMetadata converts / reads a payload into a Metadata payload.
func ReadMetadata(payload warc.IPayload) (*Metadata, error) {
	data, err := payload.Bytes()
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{}
	err = json.Unmarshal(data, metadata)
	return metadata, err
}

// Write writes the payload to a stream.
func (metadata *Metadata) Write(writer io.Writer) (int, error) {
	bytes, err := metadata.Bytes()
	if err != nil {
		return 0, err
	}

	return writer.Write(bytes)
}

// Bytes returns the byte representation of the payload.
func (metadata *Metadata) Bytes() ([]byte, error) {
	return json.Marshal(metadata)
}

// String converts the payload into a string.
func (metadata *Metadata) String() (string, error) {
	bytes, err := metadata.Bytes()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Reader returns a reader for the data.
func (metadata *Metadata) Reader() (io.Reader, error) {
	data, err := metadata.Bytes()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}

func (metadata *Metadata) Length() (uint64, error) {
	data, err := metadata.Bytes()
	if err != nil {
		return 0, err
	}

	return uint64(len(data)), nil
}
