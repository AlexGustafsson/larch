package worker

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
	"github.com/google/uuid"

	urlpkg "net/url"
)

// TODO: Naming
type RemoteWorker struct {
	JobRequests chan JobRequest
}

type Scheduler struct {
	mutex    sync.Mutex
	requests chan JobRequest
	// NOTE: No reason for these to persist - upon restart, simply reschedule jobs
	// and handle them anew.
	inflight       map[string]Job
	secret         []byte
	indexer        indexers.Indexer
	libraryReaders map[string]libraries.LibraryReader
	libraryWriters map[string]libraries.LibraryWriter
}

func NewScheduler(indexer indexers.Indexer, libraryReaders map[string]libraries.LibraryReader, libraryWriters map[string]libraries.LibraryWriter) *Scheduler {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		panic(err)
	}

	s := &Scheduler{
		requests:       make(chan JobRequest, 32),
		inflight:       make(map[string]Job),
		secret:         secret[:],
		indexer:        indexer,
		libraryReaders: libraryReaders,
		libraryWriters: libraryWriters,
	}

	return s
}

func (s *Scheduler) UpdateJob(ctx context.Context, job Job) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// TODO: E-Tag?
	s.inflight[job.ID] = job

	// TODO: Debounce
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		library, ok := s.libraryReaders[job.Library]
		if !ok {
			slog.Warn("Failed to index snapshot after job completed", slog.String("error", "no such library"))
			return
		}

		snapshotReader, err := library.ReadSnapshot(ctx, job.Origin, job.SnapshotID)
		if err != nil {
			slog.Warn("Failed to index snapshot after job completed", slog.Any("error", err))
			return
		}

		err = s.indexer.IndexSnapshot(context.Background(), job.Library, job.Origin, job.SnapshotID, snapshotReader)
		if err != nil {
			slog.Warn("Failed to index snapshot after job completed", slog.Any("error", err))
			return
		}

		slog.Debug("Successfully indexed snapshot after job completion")
	}()

	return nil
}

type GetJobOptions struct {
}

func (s *Scheduler) GetJobRequest(ctx context.Context, options *GetJobOptions) (*JobRequest, error) {
	// TODO: Could be a sync.Cond var, which would allow easier filter of jobs -
	// if not accepted, simply loop again
	slog.Debug("Waiting for job request")
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case job, ok := <-s.requests:
		if !ok {
			return nil, fmt.Errorf("closed")
		}

		return &job, nil
	}
}

// TODO: Support multiple libraries? What's the use case?
func (s *Scheduler) ScheduleSnapshot(ctx context.Context, url string, strategy *Strategy) error {
	u, err := urlpkg.Parse(url)
	if err != nil {
		return err
	}

	origin := u.Host
	snapshotID := strconv.FormatInt(time.Now().UnixMilli(), 10)

	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	id := uuid.String()

	library, ok := s.libraryWriters[strategy.Library]
	if !ok {
		return fmt.Errorf("no such library")
	}

	snapshotWriter, err := library.WriteSnapshot(ctx, origin, snapshotID)
	if err != nil {
		return err
	}

	// TODO: Include all jobs / "provenance"?
	err = snapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
		ContentType: "application/vnd.larch.snapshot.manifest.v1+json",
		Digest:      "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Size:        0,
		Annotations: map[string]string{
			"larch.snapshot.url":  url,
			"larch.snapshot.date": time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		snapshotWriter.Close()
		return err
	}

	if err := snapshotWriter.Close(); err != nil {
		return err
	}

	for _, archiver := range strategy.Archivers {
		request := JobRequest{
			Token:    "", // TODO: JWT which points to snapshot and everything?
			Archiver: archiver,
			Job: Job{
				ID:      id,
				Library: strategy.Library,
				// TODO: Once this has expired, both parties understand that the job
				// will be assumed abandoned and will be re-requested again.
				// TODO: Match this with the token, so no further requests can be made
				// TODO: Some of this time may be consumed before the worker even gets
				// the message, whilst the request is in the queue?
				Deadline:   time.Now().Add(30 * time.Minute),
				URL:        url,
				Origin:     origin,
				SnapshotID: snapshotID,
				Status:     "requested",
				Requested:  time.Now(),
			},
		}

		slog.Debug("Requesting job", slog.String("origin", origin), slog.String("snapshotId", snapshotID))
		s.mutex.Lock()
		s.inflight[request.Job.ID] = request.Job
		s.mutex.Unlock()

		// TODO: Should these be persisted instead of just a channel?
		// Could then be polled / initially built from a stateful source and then
		// event-driven
		s.requests <- request
	}

	return nil
}
