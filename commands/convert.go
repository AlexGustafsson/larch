package commands

import (
	"bufio"
	"fmt"
	"os"

	"github.com/AlexGustafsson/larch/formats/directory"
	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/urfave/cli/v2"
)

func convertCommand(context *cli.Context) error {
	inputPath := context.String("input")
	if inputPath == "" {
		return fmt.Errorf("Expected input path")
	}

	inputFormat := context.String("input-format")
	if inputFormat != "warc" && inputFormat != "warc.gz" {
		return fmt.Errorf("Expected input format to be given and one of 'warc', 'warc.gz'")
	}

	outputPath := context.String("output")
	if outputPath == "" {
		return fmt.Errorf("Expected output path to be given")
	}

	outputFormat := context.String("output-format")
	if outputFormat != "dir" && outputFormat != "tar" && outputFormat != "tgz" {
		return fmt.Errorf("Expected output format to be given and one of 'dir', 'tar', 'tgz'")
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("Unable to open input file: %v", err)
	}

	reader := bufio.NewReader(file)
	archive, err := warc.Read(reader, inputFormat == "warc.gz")
	if err != nil {
		return err
	}

	switch outputFormat {
	case "dir":
		err = directory.Marshal(archive, outputPath)
	case "tar":
		err = directory.MarshalTar(archive, outputPath, false)
	case "tgz":
		err = directory.MarshalTar(archive, outputPath, true)
	default:
		return fmt.Errorf("Format %s not supported", outputFormat)
	}

	if err != nil {
		return err
	}

	return nil
}
