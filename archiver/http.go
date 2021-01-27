package archiver

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

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

	requestRecord, err := createHTTPRequestRecord(request)
	if err != nil {
		return nil, nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}

	responseRecord, err := createHTTPResponseRecord(response)
	if err != nil {
		return nil, nil, err
	}

	return requestRecord, responseRecord, nil
}

func createHTTPRequestRecord(request *http.Request) (*warc.Record, error) {
	id, err := warc.CreateID()
	if err != nil {
		return nil, err
	}

	serializedRequest := new(bytes.Buffer)
	request.Write(serializedRequest)
	data := serializedRequest.Bytes()

	// TODO: Handle WARC-Concurrent-To to signify relationship?
	record := &warc.Record{
		Header: &warc.Header{
			Type:          warc.TypeRequest,
			TargetURI:     request.URL.String(),
			RecordID:      id,
			Date:          time.Now(),
			ContentType:   "application/http;msgtype=request",
			ContentLength: uint64(len(data)),
		},
		Payload: &warc.Payload{
			Data:   data,
			Length: uint64(len(data)),
		},
	}

	return record, nil
}

func createHTTPResponseRecord(response *http.Response) (*warc.Record, error) {
	id, err := warc.CreateID()
	if err != nil {
		return nil, err
	}

	serializedRequest := new(bytes.Buffer)
	response.Write(serializedRequest)
	data := serializedRequest.Bytes()

	record := &warc.Record{
		Header: &warc.Header{
			Type:          warc.TypeResponse,
			TargetURI:     response.Request.URL.String(),
			RecordID:      id,
			Date:          time.Now(),
			ContentType:   "application/http;msgtype=response",
			ContentLength: uint64(len(data)),
		},
		Payload: &warc.Payload{
			Data:   data,
			Length: uint64(len(data)),
		},
	}

	return record, nil
}
