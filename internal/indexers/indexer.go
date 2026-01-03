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
	ListSnapshots(context.Context, *ListSnapshotsOptions) ([]Snapshot, error)
	GetSnapshot(context.Context, string, string) (*Snapshot, error)
	GetArtifact(context.Context, string, string, string) (*Artifact, error)
	GetBlob(context.Context, string) (*Blob, error)
}

type ListSnapshotsOptions struct {
	Origin string
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
	Annotations     map[string]string
}

type Blob struct {
	ContentType     string
	ContentEncoding string
	Digest          string
	Size            int64
}
