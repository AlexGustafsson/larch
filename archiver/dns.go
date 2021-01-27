package archiver

import (
	"bytes"
	"fmt"
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/miekg/dns"
)

// CreateLookupEntry looks up a hostname's A record and creates the record.
func (archiver *Archiver) CreateLookupEntry(url *url.URL) error {
	if url.Host == "" {
		return fmt.Errorf("Invalid URL, likely missing scheme")
	}

	start := time.Now()
	client := dns.Client{}
	message := dns.Msg{}
	message.SetQuestion(url.Host+".", dns.TypeA)
	response, _, err := client.Exchange(&message, fmt.Sprintf("%s:%d", archiver.ResolverAddress, archiver.ResolverPort))
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	buffer := new(bytes.Buffer)
	for _, answer := range response.Answer {
		if record, ok := answer.(*dns.A); ok {
			fmt.Fprintf(buffer, "%s.\t%d\tIN\tA\t%s\n", url.Host, answer.Header().Ttl, record.A.String())
		}
	}

	fmt.Fprintln(buffer)
	fmt.Fprintf(buffer, ";; Query time: %d msec\n", elapsed.Milliseconds())
	fmt.Fprintf(buffer, ";; SERVER: %s#%d(%s)\n", archiver.ResolverAddress, archiver.ResolverPort, archiver.ResolverAddress)
	fmt.Fprintf(buffer, ";; WHEN: %s\n", time.Now().Format("Mon Jan 02 15:04:05 MST 2006"))

	data := buffer.Bytes()

	id, err := warc.CreateID()
	if err != nil {
		return err
	}

	record := &warc.Record{
		Header: &warc.Header{
			Type:          warc.TypeResponse,
			TargetURI:     "dns:" + url.Host,
			Date:          time.Now(),
			RecordID:      id,
			ContentType:   "text/dns",
			ContentLength: uint64(len(data)),
		},
		Payload: &warc.Payload{
			Data:   data,
			Length: uint64(len(data)),
		},
	}

	archiver.File.Records = append(archiver.File.Records, record)
	return nil
}
