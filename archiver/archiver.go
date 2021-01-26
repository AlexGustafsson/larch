package archiver

import (
	"fmt"
	"net/url"

	"github.com/AlexGustafsson/larch/warc"
	"github.com/miekg/dns"
)

// DefaultDNSProvider ...
const DefaultDNSProvider string = "192.168.1.1:53"

// Archiver contains options for an archiver.
type Archiver struct {
	MaxDepth uint32
	file     *warc.File
}

// NewArchiver creates a new archiver following best practices.
func NewArchiver() *Archiver {
	archiver := &Archiver{
		MaxDepth: 1,
		file:     &warc.File{},
	}

	return archiver
}

// CreateLookupEntry looks up a hostname's A record and creates the record.
func (archiver *Archiver) CreateLookupEntry(url *url.URL) error {
	client := dns.Client{}
	message := dns.Msg{}
	message.SetQuestion(url.Host+".", dns.TypeA)
	response, _, err := client.Exchange(&message, DefaultDNSProvider)
	if err != nil {
		return err
	}

	// There was no response
	if len(response.Answer) == 0 {
		return nil
	}

	// WARC/1.0
	// WARC-Type: response
	// WARC-Target-URI: dns:rafaela.adsclasificados.com.ar
	// WARC-Date: 2011-02-25T19:39:41Z
	// WARC-IP-Address: 207.241.228.148
	// WARC-Record-ID: <urn:uuid:757646d3-1ca9-4e0f-8775-23d5f1ed00f0>
	// Content-Type: text/dns
	// Content-Length: 73

	// 20110225193941
	// rafaela.adsclasificados.com.ar.	10800	IN	A	190.183.222.21

	for _, answer := range response.Answer {
		if record, ok := answer.(*dns.A); ok {
			fmt.Printf("%s. %d IN A %s\n", url.Host, answer.Header().Ttl, record.A.String())
		}
	}

	return nil
}

// Archive archives a URL as a WARC archive.
func (archiver *Archiver) Archive(url *url.URL) (*warc.File, error) {
	return archiver.file, nil
}
