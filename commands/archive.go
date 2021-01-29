package commands

import (
	"fmt"
	"net/url"
	"os"

	"github.com/AlexGustafsson/larch/archiver"
	"github.com/urfave/cli/v2"
)

func archiveCommand(context *cli.Context) error {
	headersOnly := context.Bool("headers-only")
	parallelism := context.Uint("parallelism")
	if parallelism < 1 {
		return fmt.Errorf("Expected a parallelism value of at least 1")
	}

	rawURLs := context.StringSlice("url")
	if len(rawURLs) == 0 {
		return fmt.Errorf("No URL given")
	}

	parsedURLs := make([]*url.URL, 0)
	for _, rawURL := range rawURLs {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("Got bad URL: %v", err)
		}

		if parsedURL.Host == "" {
			return fmt.Errorf("Got bad URL: %s", parsedURL)
		}

		parsedURLs = append(parsedURLs, parsedURL)
	}

	archiver := archiver.NewArchiver(parallelism)
	file, err := archiver.Archive(parsedURLs...)
	if err != nil {
		return err
	}

	if headersOnly {
		for _, record := range file.Records {
			record.Header.Write(os.Stdout)
		}
	} else {
		file.Write(os.Stdout)
	}

	return nil
}
