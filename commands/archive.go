package commands

import (
	"fmt"
	"net/url"

	"github.com/AlexGustafsson/larch/archiver"
	"github.com/urfave/cli/v2"
)

func archiveCommand(context *cli.Context) error {
	rawURL := context.String("url")
	if rawURL == "" {
		return fmt.Errorf("No path given")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("Got bad URL: %v", err)
	}

	archiver := archiver.NewArchiver()
	archiver.Archive(parsedURL)

	err = archiver.CreateLookupEntry(parsedURL)
	if err != nil {
		return err
	}

	return nil
}
