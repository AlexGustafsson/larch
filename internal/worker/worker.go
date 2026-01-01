package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/AlexGustafsson/larch/internal/archivers"
)

type Worker struct {
	endpoint string
	conn     *Conn
}

// TODO: For now the intent is to have a single code path for the embedded
// worker and remote workers for maintainability reasons. But we should consider
// supporting the embedded worker directly using the job channel and library,
// which could reduce memory usage by a fair bit given that there's no extra
// buffering involved
func NewWorker(ctx context.Context, endpoint string) (*Worker, error) {
	// TODO: Signal supported archivers when connecting, e.g. chrome-only worker.
	conn, err := Dial(endpoint)
	if err != nil {
		return nil, err
	}

	return &Worker{
		endpoint: endpoint,
		conn:     conn,
	}, nil
}

func (w *Worker) Work(ctx context.Context) error {
	client := &Client{
		Endpoint: w.endpoint,
	}

	for {
		requests, err := client.GetJobRequests(ctx)
		if err != nil {
			return err
		}

		for {
			select {
			case request, ok := <-requests:
				if !ok {
					// Wait before reconnecting
					select {
					case <-time.After(5 * time.Second):
						break
					case <-ctx.Done():
						return ctx.Err()
					}
				}

				err := w.work(ctx, request)
				if err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (w *Worker) work(ctx context.Context, request JobRequest) error {
	job := request.Job
	job.Accepted = time.Now()
	job.Status = "accepted"

	client := &Client{
		Endpoint: w.endpoint,
		Token:    request.Token,
	}

	err := client.UpdateJob(ctx, job)
	if err != nil {
		return err
	}

	// TODO: Initialize archiver based on config
	var archiver archivers.Archiver

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
	// TODO: Wait for current Work() call to exit?
	return w.conn.Close()
}

func (w *Worker) Close() error {
	return w.conn.Close()
}
