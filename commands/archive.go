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
	compress := context.Bool("compress")

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

	output := context.String("output")
	var outputFile *os.File
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("Unable to create output file: %v", err)
		}
		outputFile = file
	} else {
		outputFile = os.Stdout
	}

	archiver := archiver.NewArchiver(parallelism)
	file, err := archiver.Archive(parsedURLs...)
	if err != nil {
		return err
	}

	if headersOnly {
		file.WriteHeaders(outputFile, compress)
	} else {
		file.Write(outputFile, compress)
	}

	return nil
}
