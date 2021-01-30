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
		},
	},
}
