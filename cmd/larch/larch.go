package main

import (
	"context"
	"log/slog"
	"time"

	"net/http"

	"github.com/AlexGustafsson/larch/internal/api"
	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries/disk"
	"github.com/AlexGustafsson/larch/internal/worker"
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

	scheduler := worker.NewScheduler()

	webMux := http.NewServeMux()

	webMux.Handle("/api/v1/", api.NewServer(index, library))

	webServer := http.Server{
		Addr:    ":8080",
		Handler: webMux,
	}

	workerServer := http.Server{
		Addr:    ":8081",
		Handler: worker.NewAPI(scheduler, library),
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

	// Serve worker API
	wg.Go(func() error {
		err := workerServer.ListenAndServe()
		if err != http.ErrServerClosed && err != nil {
			return err
		}

		return nil
	})

	// TODO: Make "included" worker optional
	wg.Go(func() error {
		w, err := worker.NewWorker(ctx, "http://localhost:8081")
		if err != nil {
			return err
		}

		for {
			// TODO: Take timeouts from scheduler through the API instead?
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			err := w.Work(ctx)
			cancel()
			if err != nil {
				slog.Warn("Worker failed to process", slog.Any("error", err))
			}
		}
	})

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
