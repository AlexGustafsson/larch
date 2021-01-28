package jobs

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/archiver/pipeline"
	"github.com/AlexGustafsson/larch/formats/warc"
)

// HTTPRequestPayload is the payload of a HTTP request.
type HTTPRequestPayload struct {
	warc.RawPayload
	Request *http.Request
}

// HTTPResponsePayload is the payload of a HTTP response.
type HTTPResponsePayload struct {
	warc.RawPayload
	Response *http.Response
}

// CreateHTTPJob performs a HTTP GET request and returns the corresponding records.
func CreateHTTPJob(url *url.URL, userAgent string) *pipeline.Job {
	perform := func(job *pipeline.Job) ([]*warc.Record, error) {
		client := &http.Client{}

		request, err := http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return nil, err
		}

		request.Header.Add("User-Agent", userAgent)
		request.Header.Add("Accept", "*/*")

		requestRecord, err := newHTTPRequestRecord(request)
		if err != nil {
			return nil, err
		}

		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		responseRecord, err := newHTTPResponseRecord(response)
		if err != nil {
			return nil, err
		}

		return []*warc.Record{requestRecord, responseRecord}, nil
	}

	return pipeline.NewJob("HTTP", "Fetches A HTTP resource", perform)
}

// newHTTPRequestRecord creates a HTTP request record.
func newHTTPRequestRecord(request *http.Request) (*warc.Record, error) {
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
		Payload: &HTTPRequestPayload{
			RawPayload: warc.RawPayload{
				Data:   data,
				Length: uint64(len(data)),
			},
			Request: request,
		},
	}

	return record, nil
}

// newHTTPResponseRecord creates a HTTP response record.
func newHTTPResponseRecord(response *http.Response) (*warc.Record, error) {
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
		Payload: &HTTPResponsePayload{
			RawPayload: warc.RawPayload{
				Data:   data,
				Length: uint64(len(data)),
			},
			Response: response,
		},
	}

	return record, nil
}
