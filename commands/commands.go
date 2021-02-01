package commands

import (
	"github.com/urfave/cli/v2"
)

// Commands contains all commands of the application
var Commands = []*cli.Command{
	{
		Name:   "version",
		Usage:  "Show the application's version",
		Action: versionCommand,
	},
	{
		Name:   "webarchive",
		Usage:  "Work with WebArchives",
		Action: webArchiveCommand,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Path to read",
				Value: "",
			},
		},
	},
	{
		Name:   "warc",
		Usage:  "Work with WARCs",
		Action: warcCommand,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Path to read",
				Value: "",
			},
			&cli.BoolFlag{
				Name:    "compressed",
				Aliases: []string{"c"},
				Usage:   "Whether or not the WARC file is compressed using gzip.",
				Value:   false,
			},
		},
	},
	{
		Name:   "archive",
		Usage:  "Archive sites",
		Action: archiveCommand,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "url",
				Usage: "URL to a resource to archive. May be used more than once.",
			},
			&cli.BoolFlag{
				Name:  "headers-only",
				Usage: "Only print headers for the resulting WARC file",
				Value: false,
			},
			&cli.UintFlag{
				Name:  "parallelism",
				Usage: "The number of concurrent jobs to perform",
				Value: 5,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "The path to write the output too. If output is not specified, stdout will be used.",
			},
			&cli.BoolFlag{
				Name:    "compress",
				Aliases: []string{"c"},
				Usage:   "Whether or not to compress the WARC using gzip.",
				Value:   false,
			},
		},
	},
	{
		Name:   "convert",
		Usage:  "Convert between formats",
		Action: convertCommand,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Input path",
				Value:   "",
			},
			&cli.StringFlag{
				Name:  "input-format",
				Usage: "Input format. One of 'warc', 'warc.gz'",
				Value: "warc",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output path",
				Value:   "",
			},
			&cli.StringFlag{
				Name:  "output-format",
				Usage: "Output file format. One of 'dir', 'tar', 'tgz'",
				Value: "",
			},
		},
	},
	{
		Name:      "serve",
		Usage:     "Serve a WARC archive",
		Action:    serveCommand,
		ArgsUsage: "<path to archive>",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port to listen on",
				Value:   8080,
			},
			&cli.StringFlag{
				Name:  "address",
				Usage: "Address to listen on",
				Value: "",
			},
			&cli.BoolFlag{
				Name:  "no-interface",
				Usage: "Disable the Larch interface",
				Value: false,
			},
		},
	},
}
