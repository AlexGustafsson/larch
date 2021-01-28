package records

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/miekg/dns"
)

// NewDNSRecord creates a DNS response record.
func NewDNSRecord(target *url.URL, answer []dns.RR, elapsed time.Duration, resolverAddress net.IP, resolverPort uint16) (*warc.Record, error) {
	buffer := new(bytes.Buffer)
	for _, answer := range answer {
		if record, ok := answer.(*dns.A); ok {
			fmt.Fprintf(buffer, "%s.\t%d\tIN\tA\t%s\n", target.Host, answer.Header().Ttl, record.A.String())
		}
	}

	fmt.Fprintln(buffer)
	fmt.Fprintf(buffer, ";; Query time: %d msec\n", elapsed.Milliseconds())
	fmt.Fprintf(buffer, ";; SERVER: %s#%d(%s)\n", resolverAddress, resolverPort, resolverAddress)
	fmt.Fprintf(buffer, ";; WHEN: %s\n", time.Now().Format("Mon Jan 02 15:04:05 MST 2006"))

	data := buffer.Bytes()

	id, err := warc.CreateID()
	if err != nil {
		return nil, err
	}

	record := &warc.Record{
		Header: &warc.Header{
			Type:          warc.TypeResponse,
			TargetURI:     "dns:" + target.Host,
			Date:          time.Now(),
			RecordID:      id,
			ContentType:   "text/dns",
			ContentLength: uint64(len(data)),
		},
		Payload: &warc.RawPayload{
			Data:   data,
			Length: uint64(len(data)),
		},
	}

	return record, nil
}
