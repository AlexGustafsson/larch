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
}

func NewInMemoryIndex() *InMemoryIndex {
	return &InMemoryIndex{
		snapshots: make(map[string]Snapshot),
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
	url := index.Manifests[0].Layers[0].Annotations["larch.snapshot.url"]
	date, _ := time.Parse(time.RFC3339, index.Manifests[0].Layers[0].Annotations["larch.snapshot.date"])

	snapshot := Snapshot{
		URL:       url,
		Date:      date,
		Artifacts: make([]Artifact, 0),
	}

	for _, manifest := range index.Manifests {
		for _, layer := range manifest.Layers {
			snapshot.Artifacts = append(snapshot.Artifacts, Artifact{
				Type:   layer.MediaType,
				Digest: layer.Digest,
				Size:   layer.Size,
			})
		}
	}

	i.snapshots[origin+"/"+id] = snapshot
	return nil
}

// ListSnapshots implements Indexer.
func (i *InMemoryIndex) ListSnapshots(context.Context) ([]Snapshot, error) {
	return slices.Collect(maps.Values(i.snapshots)), nil
}
