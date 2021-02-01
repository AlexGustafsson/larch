package commands

import (
	"bufio"
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

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("Unable to open archive: %v", err)
	}

	reader := bufio.NewReader(file)
	archive, err := warc.Read(reader, compressed)
	if err != nil {
		return err
	}

	server := server.NewServer(archive)
	server.EnableInterface = enableInterface

	server.Start(address, port)

	return nil
}
