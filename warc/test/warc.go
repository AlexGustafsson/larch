package test

import (
	"strings"
	"testing"

	"github.com/AlexGustafsson/larch/warc"
)

func testWARCRecord(t *testing.T) {
	const raw = `
	WARC/1.0
	WARC-Type: warcinfo
	WARC-Date: 2006-09-19T17:20:14Z
	WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
	Content-Type: application/warc-fields
	Content-Length: 0`

	reader := strings.NewReader(raw)
	record, err := warc.ReadRecord(reader)
	if err != nil {
		t.Error(err)
	}

	if record.Header.ContentLength != 0 {
		t.Errorf("Expected Content-Length to be 0, was %v", record.Header.ContentLength)
	}

	if record.Header.ContentType != "application/warc-fields" {
		t.Errorf("Expected Content-Type to be application/warc-fields, was %v", record.Header.ContentType)
	}

	if record.Header.RecordID != "<urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>" {
		t.Errorf("Expected Content-Type to be <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>, was %v", record.Header.RecordID)
	}

	if record.Header.Date.String() != "2006-09-19T17:20:14Z" {
		t.Errorf("Expected Content-Type to be 2006-09-19T17:20:14Z, was %v", record.Header.Date.String())
	}

	if record.Header.Type != warc.TypeInfo {
		t.Errorf("Expected Type to be %v, was %v", warc.TypeInfo, record.Header.Type)
	}

	encoded, err := record.String()
	if encoded != raw {
		t.Errorf("Record did not survive roundtrip")
	}
}
