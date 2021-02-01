package commands

import (
	"bufio"
	"fmt"
	"os"

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

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("Unable to open archive: %v", err)
	}

	reader := bufio.NewReader(file)
	archive, err := warc.Read(reader, false)
	server := server.NewServer(archive)
	server.EnableInterface = enableInterface

	server.Start(address, port)

	return nil
}
