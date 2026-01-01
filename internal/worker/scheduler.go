package worker

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/libraries"

	urlpkg "net/url"
)

// TODO: Naming
type RemoteWorker struct {
	JobRequests chan JobRequest
}

type Scheduler struct {
	mutex    sync.Mutex
	workers  map[string]RemoteWorker
	requests chan JobRequest
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		workers:  make(map[string]RemoteWorker),
		requests: make(chan JobRequest, 32),
	}

	go s.schedule()

	return s
}

func (s *Scheduler) RegisterWorker(ctx context.Context) (string, <-chan JobRequest, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := "" // TODO

	worker := RemoteWorker{
		JobRequests: make(chan JobRequest),
	}

	s.workers[id] = worker
	return id, worker.JobRequests, nil
}

func (s *Scheduler) UnregisterWorker(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.workers, id)

	return nil
}

// TODO: Support multiple libraries? What's the use case?
func (s *Scheduler) ScheduleSnapshot(ctx context.Context, url string, archivers []archivers.Archiver, library libraries.LibraryWriter) error {
	u, err := urlpkg.Parse(url)
	if err != nil {
		panic(err)
	}

	origin := u.Host
	id := strconv.FormatInt(time.Now().UnixMilli(), 10)

	snapshotWriter, err := library.WriteSnapshot(ctx, origin, id)
	if err != nil {
		return err
	}

	// TODO: Include all jobs / "provenance"?
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
		return err
	}

	if err := snapshotWriter.Close(); err != nil {
		return err
	}

	for _, archiver := range archivers {
		// TODO
		_ = archiver
		request := JobRequest{
			Token: "", // TODO: JWT which points to snapshot and everything?
			Job: Job{
				ID: "", // TODO
			},
		}

		// TODO: Should these be persisted instead of just a channel?
		// Could then be polled / initially built from a stateful source and then
		// event-driven
		s.requests <- request
	}

	return nil
}

func (s *Scheduler) schedule() {
	// TODO: Should these be persisted instead of just a channel?
	// Could then be polled / initially built from a stateful source and then
	// event-driven
	for request := range s.requests {
		// TODO: Implement scheduling algorithm (score workers, pick best worker)
		for _, worker := range s.workers {
			worker.JobRequests <- request
			break
		}
	}
}
