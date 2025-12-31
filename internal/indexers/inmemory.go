package indexers

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ Indexer = (*InMemoryIndex)(nil)

type InMemoryIndex struct {
	snapshots map[string]Snapshot
	artifacts map[string]Artifact
}

func NewInMemoryIndex() *InMemoryIndex {
	return &InMemoryIndex{
		snapshots: make(map[string]Snapshot),
		artifacts: make(map[string]Artifact),
	}
}

// IndexLibrary implements Indexer.
func (i *InMemoryIndex) IndexLibrary(ctx context.Context, libraryReader libraries.LibraryReader) error {
	origins, err := libraryReader.GetOrigins(ctx)
	if err != nil {
		return err
	}

	for _, origin := range origins {
		snapshots, err := libraryReader.GetSnapshots(ctx, origin)
		if err != nil {
			return err
		}

		for _, id := range snapshots {
			snapshotReader, err := libraryReader.ReadSnapshot(ctx, origin, id)
			if err != nil {
				return err
			}

			if err := i.IndexSnapshot(ctx, origin, id, snapshotReader); err != nil {
				return err
			}
		}
	}

	return nil
}

// IndexSnapshot implements Indexer.
func (i *InMemoryIndex) IndexSnapshot(ctx context.Context, origin string, id string, snapshotReader libraries.SnapshotReader) error {
	index := snapshotReader.Index()

	// TODO: Fault tolerance
	url := index.Artifacts[0].Annotations["larch.snapshot.url"]
	date, _ := time.Parse(time.RFC3339, index.Artifacts[0].Annotations["larch.snapshot.date"])

	snapshot := Snapshot{
		URL:       url,
		Origin:    origin,
		ID:        id,
		Date:      date,
		Artifacts: make([]Artifact, 0),
	}

	for _, manifest := range index.Artifacts {
		artifact := Artifact{
			ContentType:     manifest.ContentType,
			ContentEncoding: manifest.ContentEncoding,
			Digest:          manifest.Digest,
			Size:            manifest.Size,
		}

		snapshot.Artifacts = append(snapshot.Artifacts, artifact)
		i.artifacts[artifact.Digest] = artifact
	}

	i.snapshots[origin+"/"+id] = snapshot
	return nil
}

// ListSnapshots implements Indexer.
func (i *InMemoryIndex) ListSnapshots(context.Context) ([]Snapshot, error) {
	return slices.Collect(maps.Values(i.snapshots)), nil
}

func (i *InMemoryIndex) GetArtifact(ctx context.Context, digest string) (*Artifact, error) {
	artifact, ok := i.artifacts[digest]
	if !ok {
		return nil, ErrNotFound
	}

	return &artifact, nil
}
