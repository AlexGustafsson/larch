package archiver

import (
	"fmt"
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/archiver/records"
	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/miekg/dns"
)

// FetchDNSRecord fetches a hostname's A record and creates the record.
func (archiver *Archiver) FetchDNSRecord(url *url.URL) (*warc.Record, error) {
	if url.Host == "" {
		return nil, fmt.Errorf("Invalid URL, likely missing scheme")
	}

	start := time.Now()
	client := dns.Client{}
	message := dns.Msg{}
	message.SetQuestion(url.Host+".", dns.TypeA)
	response, _, err := client.Exchange(&message, fmt.Sprintf("%s:%d", archiver.ResolverAddress, archiver.ResolverPort))
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)

	record, err := records.NewDNSRecord(url, response.Answer, elapsed, archiver.ResolverAddress, archiver.ResolverPort)
	if err != nil {
		return nil, err
	}

	return record, nil
}
