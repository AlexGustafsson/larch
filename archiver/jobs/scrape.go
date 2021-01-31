package jobs

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlexGustafsson/larch/archiver/pipeline"
	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// CreateScrapeJob extracts URLs in use by the target.
func CreateScrapeJob(payload *HTTPResponsePayload) *pipeline.Job {
	perform := func(job *pipeline.Job) ([]*warc.Record, error) {
		urls := make([]string, 0)

		// Read the response from the bytes instead, effectively duplicating it
		// This circumvents the issue with reading the response's body several times
		reader := bufio.NewReader(bytes.NewReader(payload.Data))
		response, err := http.ReadResponse(reader, payload.Response.Request)
		if err != nil {
			return nil, err
		}

		// Only use the actual mime type, may contain "; charset: utf-8" or the like as well
		contentType := strings.Split(response.Header.Get("Content-Type"), "; ")[0]
		switch contentType {
		case "text/html":
			err = scrapeHTML(response, payload.Response.Request.URL, &urls)
		case "text/css":
			err = scrapeCSS(response, payload.Response.Request.URL, &urls)
		default:
			log.Debugf("Skipping scraping of unsupported type %s", contentType)
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		buffer := new(bytes.Buffer)
		writer := bufio.NewWriter(buffer)
		for _, url := range urls {
			_, err := fmt.Fprintf(writer, "%s\n", url)
			if err != nil {
				return nil, err
			}
		}
		err = writer.Flush()
		if err != nil {
			return nil, err
		}

		id, err := warc.CreateID()
		if err != nil {
			return nil, err
		}

		data := buffer.Bytes()
		record := &warc.Record{
			Header: &warc.Header{
				Type:          warc.TypeConversion,
				TargetURI:     payload.Response.Request.URL.String(),
				RecordID:      id,
				Date:          time.Now(),
				ContentType:   "<metadata://github.com/AlexGustafsson/larch/scrape.txt>",
				ContentLength: uint64(len(data)),
			},
			Payload: &warc.RawPayload{
				Data:   data,
				Length: uint64(len(data)),
			},
		}
		return []*warc.Record{record}, nil
	}

	return pipeline.NewJob("Scrape", "Extracts URLs from a HTTP response", perform)
}

func scrapeHTML(response *http.Response, source *url.URL, urls *[]string) error {
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}

	document.Find("script").Each(func(i int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if exists {
			resolved, ok := resolve(src, source)
			if ok {
				*urls = append(*urls, resolved)
			}
		}
	})

	document.Find("img").Each(func(i int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if exists {
			resolved, ok := resolve(src, source)
			if ok {
				*urls = append(*urls, resolved)
			}
		}
	})

	document.Find("link").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists {
			resolved, ok := resolve(href, source)
			if ok {
				*urls = append(*urls, resolved)
			}
		}
	})

	document.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists {
			resolved, ok := resolve(href, source)
			if ok {
				*urls = append(*urls, resolved)
			}
		}
	})

	return nil
}

func scrapeCSS(response *http.Response, source *url.URL, urls *[]string) error {
	return nil
}

func resolve(link string, source *url.URL) (string, bool) {
	if strings.HasPrefix(link, "http") {
		// Absolute URL
		parsed, err := url.Parse(link)
		if err != nil || parsed.Host == "" {
			log.Debugf("Skipping invalid URL '%s'", link)
			return "", false
		}

		return parsed.String(), true
	} else if strings.HasPrefix(link, "/") {
		// Absolute path, try to resolve based on source
		return resolve(source.Scheme+"://"+source.Host+link, source)
	} else {
		// Treat as relative path
		// TODO: Might be data URL etc.
		return resolve(source.Scheme+"://"+source.Host+"/"+link, source)
	}
}
