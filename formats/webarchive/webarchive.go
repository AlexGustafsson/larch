package webarchive

import (
	"io"

	"howett.net/plist"
)

// WebResource is a WebArchive main resource
type WebResource struct {
	Data             []byte `plist:"WebResourceData"`
	FrameName        string `plist:"WebResourceFrameName"`
	MIMEType         string `plist:"WebResourceMIMEType"`
	TextEncodingName string `plist:"WebResourceTextEncodingName"`
	URL              string `plist:"WebResourceURL"`
}

// WebArchive is a representation of the WebArchive file format
type WebArchive struct {
	MainResource WebResource   `plist:"WebMainResource"`
	SubResources []WebResource `plist:"WebSubresources"`
}

// Read parses a WebArchive stream
func Read(reader io.ReadSeeker) (*WebArchive, error) {
	decoder := plist.NewDecoder(reader)
	archive := &WebArchive{}
	err := decoder.Decode(archive)
	if err != nil {
		return nil, err
	}

	return archive, nil
}

func (archive *WebArchive) Write(writer io.Writer) error {
	data, err := plist.Marshal(archive, plist.BinaryFormat)
	if err != nil {
		return err
	}

	writer.Write(data)
	return nil
}
