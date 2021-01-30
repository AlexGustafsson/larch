package warc

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
)

// File is a WARC file containing one or more records.
type File struct {
	// Records are all the records contained within the file.
	Records []*Record
}

// ReadHeaders works like Read, but only reads the headers for an entire WARC file.
func ReadHeaders(reader *bufio.Reader, compressed bool) (*File, error) {
	if compressed {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		reader = bufio.NewReader(gzipReader)
	}

	file := &File{
		Records: make([]*Record, 0),
	}

	for {
		record, err := ReadRecordHeader(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// No record received
		if record == nil {
			break
		}

		file.Records = append(file.Records, record)
	}

	return file, nil
}

// Read reads an entire WARC file.
func Read(reader *bufio.Reader, compressed bool) (*File, error) {
	if compressed {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		reader = bufio.NewReader(gzipReader)
	}

	file := &File{
		Records: make([]*Record, 0),
	}

	for {
		record, err := ReadRecord(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// No record received
		if record == nil {
			break
		}

		file.Records = append(file.Records, record)
	}

	return file, nil
}

// Write writes the file to a stream.
func (file *File) Write(writer io.Writer, compress bool) {

	for _, record := range file.Records {
		if compress {
			// Write each record in a different gzip stream to allow for "scrubbing"
			gzipWriter := gzip.NewWriter(writer)
			record.Write(gzipWriter)
			gzipWriter.Close()
		} else {
			record.Write(writer)
		}
	}
}

// WriteHeaders write only the headers to a stream (with all payloads being empty).
func (file *File) WriteHeaders(writer io.Writer, compress bool) {
	for _, record := range file.Records {
		if compress {
			// Write each record in a different gzip stream to allow for "scrubbing"
			gzipWriter := gzip.NewWriter(writer)
			record.WriteHeader(gzipWriter)
			gzipWriter.Close()
		} else {
			record.WriteHeader(writer)
		}
	}
}

// String converts the file into a string
func (file *File) String() string {
	buffer := new(bytes.Buffer)
	file.Write(buffer, false)
	return buffer.String()
}
