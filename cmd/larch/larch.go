package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	urlpkg "net/url"

	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/sources"
)

func main() {
	// library := &libraries.DiskLibrary{
	// 	BasePath: "data/disk",
	// }
	library, err := libraries.NewDiskLibrary("data/disk")
	if err != nil {
		panic(err)
	}

	archivers := []archivers.Archiver{
		&archivers.ChromeArchiver{},
		&archivers.ArchiveOrgArchiver{},
	}

	sources := []sources.Source{
		&sources.URLSource{
			URL: "https://google.com",
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
			u, err := urlpkg.Parse(url)
			if err != nil {
				panic(err)
			}

			snapshotWriter, err := library.WriteSnapshot(ctx, u.Host, strconv.FormatInt(time.Now().UnixMilli(), 10))
			if err != nil {
				panic(err)
			}

			err = snapshotWriter.WriteManifest(ctx, libraries.Manifest{
				MediaType: "application/vnd.larch.snapshot.manifest.v1+json",
				Layers: []libraries.Layer{
					{
						MediaType: "application/vnd.oci.empty.v1+json",
						Digest:    "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
						Size:      0,
						Annotations: map[string]string{
							"larch.snapshot.url":  url,
							"larch.snapshot.date": time.Now().Format(time.RFC3339),
						},
					},
				},
			})
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

	indexer := indexers.NewInMemoryIndex()
	if err := indexer.IndexLibrary(ctx, library); err != nil {
		panic(err)
	}

	snapshots, err := indexer.ListSnapshots(ctx)
	if err != nil {
		panic(err)
	}

	encoder := json.NewEncoder(os.Stderr)
	encoder.SetIndent("", "  ")
	encoder.Encode(snapshots)
}
