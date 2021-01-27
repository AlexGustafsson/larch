package warc

import (
	"bufio"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestWritePayload(t *testing.T) {
	payload := &RawPayload{
		Data:   []byte("hello, world!"),
		Length: 13,
	}

	raw := "hello, world!"

	formatted := payload.String()

	if diff := deep.Equal(formatted, raw); diff != nil {
		t.Error(diff)
	}
}

func TestReadPayload(t *testing.T) {
	raw := "hello, world!"

	header := &Header{
		ContentLength: 13,
		ContentType:   "resource",
	}

	reader := strings.NewReader(raw)
	bufferedReader := bufio.NewReader(reader)
	payload, err := ReadPayload(bufferedReader, header)
	if err != nil {
		t.Error(err)
		return
	}

	if diff := deep.Equal(payload.String(), raw); diff != nil {
		t.Error(diff)
	}
}
