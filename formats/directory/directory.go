package directory

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// MarshalTar marshals the WARC into a optionally compressed tar archive.
func MarshalTar(archive *warc.File, outputPath string, compress bool) error {
	outputDirectory, err := ioutil.TempDir("", "larch-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(outputDirectory)

	err = Marshal(archive, outputDirectory)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var tarWriter *tar.Writer
	if compress {
		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		tarWriter = tar.NewWriter(file)
	}
	defer tarWriter.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore links, directories etc.
		if !info.Mode().IsRegular() {
			return nil
		}

		// Create a header for the file or directory
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Remove the basepath to the archive, making unpacking relative
		header.Name = strings.TrimPrefix(strings.Replace(path, outputDirectory, "", -1), string(filepath.Separator))

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			file.Close()
			return err
		}

		file.Close()

		return nil
	}

	err = filepath.Walk(outputDirectory, walker)
	if err != nil {
		return err
	}

	return nil
}

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
		responseReader := bufio.NewReader(record.Payload.Reader())
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
	case "application/pdf":
		payload = record.Payload.Reader()
		name += ".pdf"
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
