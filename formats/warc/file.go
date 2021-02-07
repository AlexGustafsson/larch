package warc

import (
	"bytes"
	"compress/gzip"
	"io"
)

// File is a WARC file containing one or more records.
type File struct {
	// Records are all the records contained within the file.
	Records []*Record
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
