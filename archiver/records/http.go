package records

import (
	"bytes"
	"net/http"
	"time"

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

// NewHTTPRequestRecord creates a HTTP request record.
func NewHTTPRequestRecord(request *http.Request) (*warc.Record, error) {
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

// NewHTTPResponseRecord creates a HTTP response record.
func NewHTTPResponseRecord(response *http.Response) (*warc.Record, error) {
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
