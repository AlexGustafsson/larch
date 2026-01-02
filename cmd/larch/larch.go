package main

import (
	"context"
	"log/slog"
	"time"

	"net/http"

	"github.com/AlexGustafsson/larch/internal/api"
	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries/disk"
	"golang.org/x/sync/errgroup"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// library := &libraries.DiskLibrary{
	// 	BasePath: "data/disk",
	// }
	library, err := disk.NewLibrary("data/disk")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	index := indexers.NewInMemoryIndex()
	if err := index.IndexLibrary(ctx, library); err != nil {
		panic(err)
	}

	webMux := http.NewServeMux()

	webMux.Handle("/api/v1/", api.NewServer(index, library))

	webServer := http.Server{
		Addr:    ":8080",
		Handler: webMux,
	}

	var wg errgroup.Group

	// Serve API + web
	wg.Go(func() error {
		err := webServer.ListenAndServe()
		if err != http.ErrServerClosed && err != nil {
			return err
		}

		return nil
	})

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
