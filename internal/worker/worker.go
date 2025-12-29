package worker

import "context"

type Worker struct {
}

func (w *Worker) Snapshot(ctx context.Context, source *Source, strategy *Strategy) error {

}

type Job struct {
	Source   *Source
	Archiver *Archiver
}
