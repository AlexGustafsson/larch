package warc

import (
	"bufio"
	"bytes"
	"io"
)

// File is a WARC file containing one or more records.
type File struct {
	// Records are all the records contained within the file.
	Records []*Record
}

// Read reads an entire WARC file.
func Read(reader *bufio.Reader) (*File, error) {
	// scanner := bufio.NewScanner(reader)

	file := &File{
		Records: make([]*Record, 0),
	}

	for {
		record, err := ReadRecord(reader)
		if err != nil {
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
func (file *File) Write(writer io.Writer) {
	for _, record := range file.Records {
		record.Write(writer)
	}
}

// String converts the file into a string
func (file *File) String() string {
	buffer := new(bytes.Buffer)
	file.Write(buffer)
	return buffer.String()
}
