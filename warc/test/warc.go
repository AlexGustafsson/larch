package test

import (
	"io/ioutil"
	"testing"
	"github.com/AlexGustafsson/larch/warc"
)

func TestWARCRecord(t *testing.T) {
	const raw = `
	WARC/1.0
	WARC-Type: warcinfo
	WARC-Date: 2006-09-19T17:20:14Z
	WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
	Content-Type: application/warc-fields
	Content-Length: 0`

	record, err := warc.ParseRecord(raw)
	if err != nil {
		t.Error(err)
	}

	encoded, err := record.String()
	if encoded != raw {
		t.Errorf("Record did not survive roundtrip")
	}
}

func TestWARCInfoRecord(t *testing.T) {
	const raw = `
	WARC/1.0
	WARC-Type: warcinfo
	WARC-Date: 2006-09-19T17:20:14Z
	WARC-Record-ID: <urn:uuid:d7ae5c10-e6b3-4d27-967d-34780c58ba39>
	Content-Type: application/warc-fields
	Content-Length: 381

	software: Heritrix 1.12.0 http://crawler.archive.org
	hostname: crawling017.archive.org
	ip: 207.241.227.234
	isPartOf: testcrawl-20050708
	description: testcrawl with WARC output
	operator: IA_Admin
	http-header-user-agent: Mozilla/5.0 (compatible; heritrix/1.4.0 +http://crawler.archive.org)
	format: WARC file version 1.0
	conformsTo: http://www.archive.org/documents/WarcFileFormat-1.0.html
	`

	warcinfo := readResource("./warc/test/resources/warcinfo.txt")
	record, err :=
}
