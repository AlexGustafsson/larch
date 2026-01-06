package main

import (
	"context"
	"log/slog"
	"time"

	"net/http"

	"github.com/AlexGustafsson/larch/internal/api"
	"github.com/AlexGustafsson/larch/internal/config"
	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/libraries/disk"
	"github.com/AlexGustafsson/larch/internal/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	cfg, err := config.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	libraryReaders := make(map[string]libraries.LibraryReader)
	libraryWriters := make(map[string]libraries.LibraryWriter)
	for libraryID, library := range cfg.Libraries {
		switch library.Type {
		case "disk":
			var options config.DiskLibraryOptions
			if err := library.Options.As(&options); err != nil {
				panic(err)
			}

			// TODO: Path relative to config file
			lib, err := disk.NewLibrary(options.Path)
			if err != nil {
				panic(err)
			}

			libraryReaders[libraryID] = lib
			if !options.ReadOnly {
				libraryWriters[libraryID] = lib
			}
		}
	}

	strategies := make(map[string]worker.Strategy)
	for strategyID, strategy := range cfg.Strategies {
		archivers := make([]worker.Archiver, 0)
		for _, archiver := range strategy.Archivers {
			switch archiver.Type {
			case "archive.org":
				archivers = append(archivers, worker.Archiver{
					ArchiveOrgArchiver: &worker.ArchiveOrgArchiver{},
				})
			case "chrome":
				var options config.ChromeArchiverOptions
				if err := archiver.Options.As(&options); err != nil {
					panic(err)
				}

				resolutions := make([]worker.Resolution, 0)
				for _, resolution := range options.Screenshot.Resolutions {
					resolutions = append(resolutions, worker.Resolution(resolution))
				}

				archivers = append(archivers, worker.Archiver{
					ChromeArchiver: &worker.ChromeArchiver{
						SavePDF:               options.PDF.Enabled,
						SaveSinglefile:        options.Singlefile.Enabled,
						ScreenshotResolutions: resolutions,
					},
				})
			case "opengraph":
				archivers = append(archivers, worker.Archiver{
					OpenGraphArchiver: &worker.OpenGraphArchiver{},
				})
			}
		}
		strategies[strategyID] = worker.Strategy{
			Library:   strategy.Library,
			Archivers: archivers,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	index := indexers.NewInMemoryIndex()
	for libraryID, library := range libraryReaders {
		if err := index.IndexLibrary(ctx, libraryID, library); err != nil {
			panic(err)
		}
	}

	scheduler := worker.NewScheduler(index, libraryReaders, libraryWriters)

	webMux := http.NewServeMux()

	webMux.Handle("/api/v1/", api.NewServer(index, libraryReaders))

	webServer := http.Server{
		Addr:    ":8080",
		Handler: webMux,
	}

	workerAPI := worker.NewAPI(scheduler, libraryWriters)

	workerServer := http.Server{
		Addr:    ":8081",
		Handler: workerAPI,
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

	// Run a default worker
	wg.Go(func() error {
		worker := worker.NewWorker("http://localhost:8081")
		return worker.Work(context.Background())
	})

	// Trigger test job
	for _, source := range cfg.Sources {
		switch source.Type {
		case "url":
			var options config.URLSourceOptions
			if err := source.Options.As(&options); err != nil {
				panic(err)
			}

			strategy, ok := strategies[source.Strategy]
			if !ok {
				panic("invalid strategy")
			}

			err := scheduler.ScheduleSnapshot(context.Background(), options.URL, &strategy)
			if err != nil {
				panic(err)
			}
		}
	}

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
