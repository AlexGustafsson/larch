package commands

import (
	"fmt"
	"os"

	"github.com/AlexGustafsson/larch/plugin"
	"github.com/urfave/cli/v2"
)

func webArchiveCommand(context *cli.Context) error {
	manager := plugin.NewManager()
	err := manager.RegisterFormatter("webarchive", "./build/plugins/webarchive")
	if err != nil {
		return err
	}

	path := context.String("path")
	if path == "" {
		return fmt.Errorf("No path given")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	webarchive := manager.Formatters["webarchive"]

	archive, err := webarchive.Read(file)
	if err != nil {
		return err
	}

	fmt.Printf("URL: %s\n", archive.MainResource.URL)
	for _, resource := range archive.SubResources {
		fmt.Printf("URL: %s\n", resource.URL)
	}

	return nil
}
