package jobs

import (
	"net/http"
	"strings"

	"github.com/AlexGustafsson/larch/archiver/pipeline"
	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// CreateScrapeJob extracts URLs in use by the target.
// TODO: Parse all paths and create URLs to return instead
func CreateScrapeJob(payload *HTTPResponsePayload) *pipeline.Job {
	perform := func(job *pipeline.Job) ([]*warc.Record, error) {
		urls := make([]string, 0)

		// Only use the actual mime type, may contain "; charset: utf-8" or the like as well
		contentType := strings.Split(payload.Response.Header.Get("Content-Type"), "; ")[0]
		var err error
		switch contentType {
		case "text/html":
			err = scrapeHTML(payload.Response, &urls)
		case "text/css":
			err = scrapeCSS(payload.Response, &urls)
		default:
			log.Debugf("Skipping scraping of unsupported type %s", contentType)
		}
		if err != nil {
			return nil, err
		}

		// return urls, nil
		return nil, nil
	}

	return pipeline.NewJob("Scrape", "Extracts URLs from a HTTP response", perform)
}

func scrapeHTML(response *http.Response, urls *[]string) error {
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}

	// TODO: Parallelize
	// TODO: Extract inline CSS for processing with scrapeCSS
	document.Find("script").Each(func(i int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if exists {
			*urls = append(*urls, src)
		}
	})

	document.Find("img").Each(func(i int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if exists {
			*urls = append(*urls, src)
		}
	})

	document.Find("link").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists {
			*urls = append(*urls, href)
		}
	})

	document.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if exists {
			*urls = append(*urls, href)
		}
	})

	return nil
}

func scrapeCSS(response *http.Response, urls *[]string) error {
	// TODO
	return nil
}
