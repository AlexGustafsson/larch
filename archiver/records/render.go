package records

import (
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// NewRenderRecord creates a record for a rendering of a site.
func NewRenderRecord(url *url.URL, buffer []byte) (*warc.Record, error) {
	id, err := warc.CreateID()
	if err != nil {
		return nil, err
	}

	// TODO: Handle WARC-Corresponds-To to signify relationship?
	record := &warc.Record{
		Header: &warc.Header{
			Type:          warc.TypeConversion,
			TargetURI:     url.String(),
			RecordID:      id,
			Date:          time.Now(),
			ContentType:   "image/png",
			ContentLength: uint64(len(buffer)),
		},
		Payload: &warc.RawPayload{
			Data:   buffer,
			Length: uint64(len(buffer)),
		},
	}

	return record, nil
}
