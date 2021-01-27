package commands

import (
	"bufio"
	"fmt"
	"os"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/urfave/cli/v2"
)

func warcCommand(context *cli.Context) error {
	path := context.String("path")
	if path == "" {
		return fmt.Errorf("No path given")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	archive, err := warc.ReadHeaders(reader)
	if err != nil {
		return err
	}

	for _, record := range archive.Records {
		fmt.Printf("%v (%v) - %vB\n", record.Header.Type, record.Header.ContentType, record.Header.ContentLength)
	}

	return nil
}
