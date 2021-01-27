package commands

import (
	"fmt"
	"net/url"
	"os"

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

	if parsedURL.Host == "" {
		return fmt.Errorf("Got bad URL: %s", parsedURL)
	}

	archiver := archiver.NewArchiver()
	err = archiver.Archive(parsedURL)
	if err != nil {
		return err
	}

	archiver.File.Write(os.Stdout)

	return nil
}
