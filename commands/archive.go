package commands

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/archiver"
	"github.com/AlexGustafsson/larch/formats/directory"
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

	// archiver.File.Write(os.Stdout)
	resolvedPath, err := filepath.Abs("./data/test-output")
	if err != nil {
		return err
	}

	for _, record := range archiver.File.Records {
		record.Header.Write(os.Stdout)
	}

	err = directory.Marshal(archiver.File, resolvedPath)
	if err != nil {
		return err
	}

	return nil
}
