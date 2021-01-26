package archiver

import (
	"net"
	"net/url"

	"github.com/AlexGustafsson/larch/warc"
)

// Archiver contains options for an archiver.
type Archiver struct {
	MaxDepth        uint32
	File            *warc.File
	ResolverAddress net.IP
	ResolverPort    uint16
}

// NewArchiver creates a new archiver following best practices.
func NewArchiver() *Archiver {
	archiver := &Archiver{
		MaxDepth:        1,
		File:            &warc.File{},
		ResolverAddress: net.ParseIP("192.168.1.1"),
		ResolverPort:    uint16(53),
	}

	return archiver
}

// Archive archives a URL as a WARC archive.
func (archiver *Archiver) Archive(url *url.URL) error {
	err := archiver.CreateLookupEntry(url)
	if err != nil {
		return err
	}

	return nil
}
