package main

import (
	"context"
	"strconv"
	"time"

	"net/http"
	urlpkg "net/url"

	"github.com/AlexGustafsson/larch/internal/api"
	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/libraries/disk"
	"github.com/AlexGustafsson/larch/internal/sources"
)

func main() {
	// library := &libraries.DiskLibrary{
	// 	BasePath: "data/disk",
	// }
	library, err := disk.NewLibrary("data/disk")
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

			err = snapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
				ContentType: "application/vnd.larch.snapshot.manifest.v1+json",
				Digest:      "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
				Size:        0,
				Annotations: map[string]string{
					"larch.snapshot.url":  url,
					"larch.snapshot.date": time.Now().Format(time.RFC3339),
				},
			})
			if err != nil {
				snapshotWriter.Close()
				panic(err)
			}

			for _, archiver := range archivers {
				err := archiver.Archive(ctx, snapshotWriter, url)
				if err != nil {
					snapshotWriter.Close()
					panic(err)
				}
			}

			if err := snapshotWriter.Close(); err != nil {
				panic(err)
			}
		}
	}

	index := indexers.NewInMemoryIndex()
	if err := index.IndexLibrary(ctx, library); err != nil {
		panic(err)
	}

	apiServer := api.NewServer(index, library)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/", apiServer)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != http.ErrServerClosed && err != nil {
		panic(err)
	}
}
