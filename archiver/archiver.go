package archiver

import (
	"fmt"
	"net"
	"net/url"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// Archiver contains options for an archiver.
type Archiver struct {
	MaxDepth        uint32
	File            *warc.File
	ResolverAddress net.IP
	ResolverPort    uint16
	UserAgent       string
}

// NewArchiver creates a new archiver following best practices.
func NewArchiver() *Archiver {
	archiver := &Archiver{
		MaxDepth:        1,
		File:            &warc.File{},
		ResolverAddress: net.ParseIP("192.168.1.1"),
		ResolverPort:    uint16(53),
		UserAgent:       "Larch (github.com/AlexGustafsson/larc)",
	}

	return archiver
}

// Archive archives a URL as a WARC archive.
func (archiver *Archiver) Archive(url *url.URL) error {
	err := archiver.CreateLookupEntry(url)
	if err != nil {
		return err
	}

	requestRecord, responseRecord, err := archiver.FetchRobotsTXT(url)
	if err != nil {
		return err
	}
	archiver.File.Records = append(archiver.File.Records, requestRecord)
	archiver.File.Records = append(archiver.File.Records, responseRecord)

	requestRecord, responseRecord, err = archiver.Fetch(url)
	if err != nil {
		return err
	}
	archiver.File.Records = append(archiver.File.Records, requestRecord)
	archiver.File.Records = append(archiver.File.Records, responseRecord)

	urls, err := archiver.Scrape(responseRecord.Payload.Reader())
	if err != nil {
		return err
	}
	fmt.Println(urls)

	// renderRecord, err := archiver.RenderSite(url, 100)
	// if err != nil {
	// 	return err
	// }
	// archiver.File.Records = append(archiver.File.Records, renderRecord)

	return nil
}
