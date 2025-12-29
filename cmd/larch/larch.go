package main

import (
	"context"
	"strconv"
	"time"

	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/sources"
)

func main() {
	library := &libraries.DiskLibrary{
		BasePath: "data",
	}

	archivers := []archivers.Archiver{
		&archivers.ChromeArchiver{},
		&archivers.ArchiveOrgArchiver{},
	}

	sources := []sources.Source{
		&sources.URLSource{
			URL: "https://google.se",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for _, source := range sources {
		urls, err := source.URLs(ctx)
		if err != nil {
			panic(err)
		}

		for _, url := range urls {
			origin := "github.com"
			snapshotWriter, err := library.OpenSnapshot(ctx, origin+"/"+strconv.FormatInt(time.Now().UnixMilli(), 10))
			if err != nil {
				panic(err)
			}

			for _, archiver := range archivers {
				err := archiver.Archive(ctx, snapshotWriter, url)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
