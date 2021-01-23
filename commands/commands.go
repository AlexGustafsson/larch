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
}
