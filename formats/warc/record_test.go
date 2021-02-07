package warc

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestReadRecord(t *testing.T) {
	raw := `WARC/1.0
WARC-Type: warcinfo
WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
WARC-Date: 2006-09-19T19:20:14+0200
Content-Length: 11
Content-Type: text/raw
WARC-Segment-Number: 0
WARC-Segment-Total-Length: 0

hello world`
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\n", "\r\n")
	raw += "\r\n\r\n"

	reader, err := NewReader(strings.NewReader(raw), false)
	if err != nil {
		t.Error(err)
		return
	}

	record, err := reader.ReadRecord()
	if err != nil {
		t.Error(err)
		return
	}

	if record.Header == nil {
		t.Errorf("Expected Header to be set, was nil")
	}

	if record.Payload == nil {
		t.Errorf("Expected Payload to be set, was nil")
	}

	encoded, err := record.String()
	if err != nil {
		t.Error(err)
	}

	if diff := deep.Equal(encoded, raw); diff != nil {
		t.Error(diff)
	}
}

func TestReadRecordSeparate(t *testing.T) {
	raw := `WARC/1.0
WARC-Type: warcinfo
WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
WARC-Date: 2006-09-19T19:20:14+0200
Content-Length: 11
Content-Type: text/raw
WARC-Segment-Number: 0
WARC-Segment-Total-Length: 0

hello world`
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\n", "\r\n")
	raw += "\r\n\r\n"

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

	if diff := deep.Equal(header.payloadOffset, int64(233)); diff != nil {
		t.Error(diff)
	}

	payload, err := reader.ReadPayload(header)
	if err != nil {
		t.Error(err)
		return
	}

	serializedPayload, err := payload.String()
	if err != nil {
		t.Error(err)
		return
	}

	if diff := deep.Equal(serializedPayload, "hello world"); diff != nil {
		t.Error(diff)
	}
}
