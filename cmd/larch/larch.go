package main

import (
	"context"
	"log/slog"
	"time"

	"net/http"

	"github.com/AlexGustafsson/larch/internal/api"
	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/AlexGustafsson/larch/internal/libraries/disk"
	"github.com/AlexGustafsson/larch/internal/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// library := &libraries.DiskLibrary{
	// 	BasePath: "data/disk",
	// }
	diskLibrary, err := disk.NewLibrary("data/disk")
	if err != nil {
		panic(err)
	}

	diskLibrary2, err := disk.NewLibrary("data/disk2")
	if err != nil {
		panic(err)
	}

	libraryReaders := map[string]libraries.LibraryReader{
		"disk":  diskLibrary,
		"disk2": diskLibrary2,
	}

	libraryWriters := map[string]libraries.LibraryWriter{
		"disk":  diskLibrary,
		"disk2": diskLibrary2,
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
	scheduler.ScheduleSnapshot(context.Background(), "https://google.se", &worker.Strategy{
		Library: "disk",
		Archivers: []worker.Archiver{
			{
				ArchiveOrgArchiver: &worker.ArchiveOrgArchiver{},
			},
			{
				ChromeArchiver: &worker.ChromeArchiver{
					SavePDF:               true,
					SaveSinglefile:        true,
					ScreenshotResolutions: []worker.Resolution{"1280x720"},
				},
			},
		},
	})

	scheduler.ScheduleSnapshot(context.Background(), "https://google.se", &worker.Strategy{
		Library: "disk2",
		Archivers: []worker.Archiver{
			{
				ArchiveOrgArchiver: &worker.ArchiveOrgArchiver{},
			},
			{
				ChromeArchiver: &worker.ChromeArchiver{
					SavePDF:        true,
					SaveSinglefile: true,
				},
			},
		},
	})

	if err := wg.Wait(); err != nil {
		panic(err)
	}
}
