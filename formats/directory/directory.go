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
		// TODO: Parallelize
		err = marshalRecord(record, outputDirectory)
		if err != nil {
			return fmt.Errorf("Unable to marshal record: %v", err)
		}
	}

	return nil
}

func marshalRecord(record *warc.Record, outputDirectory string) error {
	// Skip empty payloads
	if record.Header.ContentLength == 0 || record.Payload == nil {
		log.Debugf("Skipping empty record payload")
		return nil
	}

	id, err := warc.ParseID(record.Header.RecordID)
	if err != nil {
		return err
	}

	var payload io.Reader
	name := id.String()
	switch record.Header.ContentType {
	case "application/http;msgtype=response":
		responseReader := bufio.NewReader(bytes.NewReader(record.Payload.Bytes()))
		response, err := http.ReadResponse(responseReader, nil)
		if err != nil {
			return fmt.Errorf("Unable to parse HTTP response: %v", err)
		}

		// For now, skip non-ok responses
		if response.StatusCode != 200 {
			log.Debugf("Skipping non-OK response (%d)", response.StatusCode)
			return nil
		}

		payload = response.Body
		name += ".html"
	case "image/png":
		payload = record.Payload.Reader()
		name += ".png"
	default:
		log.Debugf("Defaulting to raw file for content type '%s'", record.Header.ContentType)
		payload = record.Payload.Reader()
	}

	filePath := filepath.Join(outputDirectory, name)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Unable to create file: %v", err)
	}

	writtenBytes, err := io.Copy(file, payload)
	if err != nil {
		return fmt.Errorf("Unable to write to file: %v", err)
	}
	log.Debugf("Wrote %d bytes to %s", writtenBytes, filePath)

	return nil
}
