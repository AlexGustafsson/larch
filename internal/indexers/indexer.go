package indexers

import (
	"context"
	"errors"
	"time"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var (
	ErrNotFound = errors.New("not found")
)

type Indexer interface {
	IndexLibrary(context.Context, libraries.LibraryReader) error
	IndexSnapshot(context.Context, string, string, libraries.SnapshotReader) error
	ListSnapshots(context.Context) ([]Snapshot, error)
	GetArtifact(context.Context, string) (*Artifact, error)
}

type Snapshot struct {
	URL       string
	Origin    string
	ID        string
	Date      time.Time
	Artifacts []Artifact
}

type Artifact struct {
	ContentType     string
	ContentEncoding string
	Digest          string
	Size            int64
}
