package indexers

import (
	"context"
	"time"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type Indexer interface {
	IndexLibrary(context.Context, libraries.LibraryReader) error
	IndexSnapshot(context.Context, string, string, libraries.SnapshotReader) error
	ListSnapshots(context.Context) ([]Snapshot, error)
}

type Snapshot struct {
	URL       string
	Date      time.Time
	Artifacts []Artifact
}

type Artifact struct {
	Type   string
	Digest string
	Size   int64
}
