package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/urfave/cli/v2"
)

func warcCommand(context *cli.Context) error {
	path := context.String("path")
	if path == "" {
		return fmt.Errorf("No path given")
	}

	// Compressed is false by default, but reacts to the suffix of the file path
	compressed := context.Bool("compressed")
	if !compressed {
		compressed = strings.HasSuffix(path, ".gz")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	reader, err := warc.NewReader(file, compressed)
	if err != nil {
		return err
	}

	archive, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range archive.Records {
		fmt.Printf("%v (%v) - %vB\n", record.Header.Type, record.Header.ContentType, record.Header.ContentLength)
	}

	return nil
}
