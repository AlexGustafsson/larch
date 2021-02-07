package warc

import (
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestWriteHeader(t *testing.T) {
	header := Header{
		Type:          "warcinfo",
		Date:          time.Unix(1158686414, 0),
		RecordID:      "<urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>",
		ContentType:   "application/warc-fields",
		ContentLength: 0,
	}

	raw := `WARC/1.0
WARC-Type: warcinfo
WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
WARC-Date: 2006-09-19T19:20:14+0200
Content-Length: 0
Content-Type: application/warc-fields
WARC-Segment-Number: 0
WARC-Segment-Total-Length: 0`
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\n", "\r\n")
	raw += "\r\n"

	formatted, err := header.String()
	if err != nil {
		t.Error(err)
	}

	if diff := deep.Equal(formatted, raw); diff != nil {
		t.Error(diff)
	}
}

func TestReadHeader(t *testing.T) {
	raw := `WARC/1.0
WARC-Type: warcinfo
WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
WARC-Date: 2006-09-19T19:20:14+0200
Content-Length: 0
Content-Type: application/warc-fields
WARC-Segment-Number: 0
WARC-Segment-Total-Length: 0`
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\n", "\r\n")
	raw += "\r\n"

	reader, err := NewReader(strings.NewReader(raw), false)
	if err != nil {
		t.Error(err)
		return
	}

	header, err := reader.ReadHeader()
	if err != nil {
		t.Error(err)
		return
	}

	if diff := deep.Equal(header.Type, "warcinfo"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(header.Date, time.Unix(1158686414, 0)); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(header.ContentType, "application/warc-fields"); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(header.ContentLength, uint64(0)); diff != nil {
		t.Error(diff)
	}

	if diff := deep.Equal(header.payloadOffset, int64(245)); diff != nil {
		t.Error(diff)
	}

	encoded, err := header.String()
	if err != nil {
		t.Error(err)
	}

	if diff := deep.Equal(encoded, raw); diff != nil {
		t.Error(diff)
	}
}
