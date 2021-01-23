package commands

import (
	"fmt"
	"os"

	"github.com/AlexGustafsson/larch/webarchive"
	"github.com/urfave/cli/v2"
)

func webArchiveCommand(context *cli.Context) error {
	path := context.String("path")
	if path == "" {
		return fmt.Errorf("No path given")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	file2, err := os.Create(path + "2")
	if err != nil {
		return err
	}

	archive, err := webarchive.ParseWebArchive(file)
	if err != nil {
		return err
	}

	fmt.Printf("URL: %s\n", archive.MainResource.URL)
	for _, resource := range archive.SubResources {
		fmt.Printf("URL: %s\n", resource.URL)
	}

	err = archive.Write(file2)
	if err != nil {
		return err
	}

	return nil
}
