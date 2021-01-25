package warc

import (
	"testing"
	"time"

	"github.com/go-test/deep"
)

type exampleMarshalStruct struct {
	// Type is an example WARC header value.
	TypeWithAnnotation      string `warc:"WARC-Type"`
	TypeWithoutAnnotation   string
	OmittedForBeingEmpty    string `warc:"omitempty"`
	NotOmittedForBeingThere string `warc:"omitempty"`
	TimeType                time.Time
	DurationType            time.Duration
}

func TestMarshal(t *testing.T) {
	tenSeconds, _ := time.ParseDuration("10s")
	v := exampleMarshalStruct{
		TypeWithAnnotation:      "foo",
		TypeWithoutAnnotation:   "bar",
		NotOmittedForBeingThere: "hello",
		TimeType:                time.Unix(1611593130, 0),
		DurationType:            tenSeconds,
	}

	expected := "WARC-Type: foo\r\nTypeWithoutAnnotation: bar\r\nNotOmittedForBeingThere: hello\r\nTimeType: 2021-01-25T17:45:30+0100\r\nDurationType: 10000\r\n"

	data, err := Marshal(v)
	if err != nil {
		t.Error(err)
	}
	encoded := string(data)

	if diff := deep.Equal(encoded, expected); diff != nil {
		t.Error(diff)
	}
}
