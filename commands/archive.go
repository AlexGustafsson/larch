package commands

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/AlexGustafsson/larch/archiver"
	"github.com/AlexGustafsson/larch/formats/directory"
	log "github.com/sirupsen/logrus"
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

	path, err := filepath.Abs("./data/test-output")
	if err != nil {
		return err
	}

	log.Debugf("Marshalling to output directory: %s", path)
	err = directory.Marshal(archiver.File, path)
	if err != nil {
		return err
	}

	return nil
}
