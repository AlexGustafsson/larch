package archiver

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/AlexGustafsson/larch/warc"
	"github.com/gocolly/colly"
)

// Archiver contains options for an archiver.
type Archiver struct {
	MaxDepth  uint32
	File      *warc.File
	collector *colly.Collector
}

func serializeRequest(request *colly.Request) ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := bufio.NewWriter(buffer)

	fmt.Fprintf(writer, "%v %v HTTP/1.1\r\n", request.Method, request.URL.Path)

	request.Headers.Write(writer)
	writer.WriteString("\r\n")

	// Consuming the request's body means we'll need to replace it with a new reader
	if request.Body != nil {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		writer.Write(body)
		request.Body = bytes.NewReader(body)
	}

	return buffer.Bytes(), nil
}

func serializeResponse(response *colly.Response) []byte {
	buffer := new(bytes.Buffer)
	writer := bufio.NewWriter(buffer)
	fmt.Fprintf(writer, "HTTP/1.1 %v\r\n", response.StatusCode)
	response.Headers.Write(writer)
	writer.WriteString("\r\n")
	writer.Write(response.Body)
	return buffer.Bytes()
}

// NewArchiver creates a new archiver following best practices.
func NewArchiver() *Archiver {
	archiver := &Archiver{
		MaxDepth:  1,
		File:      &warc.File{},
		collector: colly.NewCollector(),
	}

	archiver.collector.Async = true
	archiver.collector.MaxDepth = int(archiver.MaxDepth)

	archiver.collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL)
		data, err := serializeRequest(request)
		if err != nil {
			return
		}

		record := &warc.Record{
			Header: &warc.Header{
				Type:          "request",
				Date:          time.Now(),
				ContentType:   "application/http;msgtype=request",
				ContentLength: uint64(len(data)),
			},
			Payload: &warc.Payload{
				Data:   data,
				Length: uint64(len(data)),
			},
		}

		archiver.File.Records = append(archiver.File.Records, record)
	})

	archiver.collector.OnResponse(func(response *colly.Response) {
		fmt.Println("Visited", response.Request.URL)
		data := serializeResponse(response)

		record := &warc.Record{
			Header: &warc.Header{
				Type:          "response",
				Date:          time.Now(),
				ContentType:   "application/http;msgtype=response",
				ContentLength: uint64(len(data)),
			},
			Payload: &warc.Payload{
				Data:   data,
				Length: uint64(len(data)),
			},
		}

		archiver.File.Records = append(archiver.File.Records, record)
	})

	return archiver
}

// Archive archives a URL as a WARC archive.
func (archiver *Archiver) Archive(url *url.URL) (*warc.File, error) {
	archiver.collector.Visit(url.String())
	archiver.collector.Wait()
	return nil, nil
}
