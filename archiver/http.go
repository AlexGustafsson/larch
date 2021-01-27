package archiver

import (
	"net/http"
	"net/url"

	"github.com/AlexGustafsson/larch/archiver/records"
	"github.com/AlexGustafsson/larch/formats/warc"
)

// Fetch performs a HTTP GET request and returns the corresponding records.
func (archiver *Archiver) Fetch(url *url.URL) (*warc.Record, *warc.Record, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	request.Header.Add("User-Agent", archiver.UserAgent)
	request.Header.Add("Accept", "*/*")

	requestRecord, err := records.NewHTTPRequestRecord(request)
	if err != nil {
		return nil, nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}

	responseRecord, err := records.NewHTTPResponseRecord(response)
	if err != nil {
		return nil, nil, err
	}

	return requestRecord, responseRecord, nil
}
