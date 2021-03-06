package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/AlexGustafsson/larch/server"
	"github.com/urfave/cli/v2"
)

func serveCommand(context *cli.Context) error {
	archivePath := context.Args().Get(0)
	if archivePath == "" {
		return fmt.Errorf("Expected path to WARC archive")
	}

	address := context.String("address")
	port := uint16(context.Uint("port"))
	enableInterface := !context.Bool("no-interface")

	// Compressed is false by default, but reacts to the suffix of the file path
	compressed := context.Bool("compressed")
	if !compressed {
		compressed = strings.HasSuffix(archivePath, ".gz")
	}

	site := context.String("site")
	if site == "" {
		return fmt.Errorf("Expected site to serve")
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("Unable to open archive: %v", err)
	}

	reader, err := warc.NewReader(file, compressed)
	if err != nil {
		return err
	}

	archive, err := reader.ReadAllHeaders()
	if err != nil {
		return err
	}

	server := server.NewServer(reader, archive, site)
	server.EnableInterface = enableInterface

	server.Start(address, port)

	return nil
}
