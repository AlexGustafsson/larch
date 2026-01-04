package worker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/archivers/chrome"
)

type Worker struct {
	endpoint string
}

// TODO: For now the intent is to have a single code path for the embedded
// worker and remote workers for maintainability reasons. But we should consider
// supporting the embedded worker directly using the job channel and library,
// which could reduce memory usage by a fair bit given that there's no extra
// buffering involved
func NewWorker(endpoint string) *Worker {
	return &Worker{
		endpoint: endpoint,
	}
}

func (w *Worker) Work(ctx context.Context) error {
	client := &Client{
		Endpoint: w.endpoint,
		Client:   http.DefaultClient,
	}

	for {
		slog.Debug("Waiting for job")
		jobRequest, err := client.GetJobRequest(ctx)
		if err != nil {
			return err
		}

		slog.Debug("Got job, executing", slog.String("jobId", jobRequest.Job.ID), slog.String("origin", jobRequest.Job.Origin), slog.String("snapshotID", jobRequest.Job.SnapshotID))
		err = w.work(ctx, jobRequest)
		if err != nil {
			return err
		}
	}
}

func (w *Worker) work(ctx context.Context, request *JobRequest) error {
	job := request.Job
	job.Accepted = time.Now()
	job.Status = "accepted"

	client := &JobClient{
		Origin:     job.Origin,
		SnapshotID: job.SnapshotID,
		Endpoint:   w.endpoint,
		Token:      request.Token,
		Client:     http.DefaultClient,
	}

	err := client.UpdateJob(ctx, job)
	if err != nil {
		return err
	}

	// TODO: Initialize archiver based on config
	var archiver archivers.Archiver
	if request.Archiver.ChromeArchiver != nil {
		screenshotResolutions := make([]chrome.Resolution, 0)
		for _, resolution := range request.Archiver.ChromeArchiver.ScreenshotResolutions {
			width, height, err := resolution.Rect()
			if err != nil {
				return err
			}

			screenshotResolutions = append(screenshotResolutions, chrome.Resolution{Width: width, Height: height})
		}

		archiver = &chrome.Archiver{
			ScreenshotResolutions: screenshotResolutions,
			SavePDF:               request.Archiver.ChromeArchiver.SavePDF,
			SaveSinglefile:        request.Archiver.ChromeArchiver.SaveSinglefile,
		}
	} else if request.Archiver.ArchiveOrgArchiver != nil {
		archiver = &archivers.ArchiveOrgArchiver{}
	} else {
		panic("invalid archiver options")
	}

	job.Started = time.Now()
	job.Status = "started"

	err = client.UpdateJob(ctx, job)
	if err != nil {
		return err
	}

	err = archiver.Archive(ctx, client, job.URL)
	job.Ended = time.Now()

	if err == nil {
		job.Status = "succeeded"
	} else {
		job.Status = "failed"
		// TODO: Don't expose all error types?
		job.Error = err.Error()
		slog.Warn("Failed to archive", slog.Any("error", err))
	}

	return client.UpdateJob(ctx, job)
}

func (w *Worker) Shutdown() error {
	panic("not implemented")
}

func (w *Worker) Close() error {

	panic("not implemented")
}
