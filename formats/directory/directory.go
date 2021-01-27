package directory

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// Marshal extracts files from a WARC to a directory.
// Expects outputDirectory to be a resolved, absolute path.
// Creates the directory if it does not exist.
func Marshal(archive *warc.File, outputDirectory string) error {
	err := os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to create output directory: %v", err)
	}

	for _, record := range archive.Records {
		err = marshalRecord(record, outputDirectory)
		if err != nil {
			return fmt.Errorf("Unable to marshal record: %v", err)
		}
	}

	return nil
}

func marshalRecord(record *warc.Record, outputDirectory string) error {
	// For now, only handle HTTP responses
	if record.Header.ContentType != "application/http;msgtype=response" {
		log.Debugf("Skipping record which was not a HTTP response (%s)", record.Header.ContentType)
		return nil
	}

	// Skip empty payloads
	if record.Payload == nil {
		log.Debugf("Skipping empty record payload")
		return nil
	}

	responseReader := bufio.NewReader(bytes.NewReader(record.Payload.Data))
	response, err := http.ReadResponse(responseReader, nil)
	if err != nil {
		return fmt.Errorf("Unable to parse HTTP response: %v", err)
	}

	// For now, skip non-ok responses
	if response.StatusCode != 200 {
		log.Debugf("Skipping non-OK response (%d)", response.StatusCode)
		return nil
	}

	filePath := filepath.Join(outputDirectory, "index.html")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Unable to create file: %v", err)
	}

	writtenBytes, err := io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("Unable to write to file: %v", err)
	}
	log.Debugf("Wrote %d bytes to %s", writtenBytes, filePath)

	return nil
}
