package indexers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ Indexer = (*InMemoryIndex)(nil)

type InMemoryIndex struct {
	snapshots map[string]Snapshot
	blobs     map[string]Blob
}

func NewInMemoryIndex() *InMemoryIndex {
	return &InMemoryIndex{
		snapshots: make(map[string]Snapshot),
		blobs:     make(map[string]Blob),
	}
}

// IndexLibrary implements Indexer.
func (i *InMemoryIndex) IndexLibrary(ctx context.Context, libraryID string, libraryReader libraries.LibraryReader) error {
	origins, err := libraryReader.GetOrigins(ctx)
	if err != nil {
		return fmt.Errorf("failed to get origins: %w", err)
	}

	for _, origin := range origins {
		snapshots, err := libraryReader.GetSnapshots(ctx, origin)
		if err != nil {
			return fmt.Errorf("failed to get snapshots for origin: %w", err)
		}

		for _, snapshotID := range snapshots {
			snapshotReader, err := libraryReader.ReadSnapshot(ctx, origin, snapshotID)
			if err != nil {
				return fmt.Errorf("failed to read snapshot: %w", err)
			}

			if err := i.IndexSnapshot(ctx, libraryID, origin, snapshotID, snapshotReader); err != nil {
				return fmt.Errorf("failed to index snapshot: %w", err)
			}
		}
	}

	return nil
}

// IndexSnapshot implements Indexer.
func (i *InMemoryIndex) IndexSnapshot(ctx context.Context, libraryID string, origin string, snapshotID string, snapshotReader libraries.SnapshotReader) error {
	index := snapshotReader.Index()

	// TODO: Fault tolerance
	url := index.Artifacts[0].Annotations["larch.snapshot.url"]
	date, _ := time.Parse(time.RFC3339, index.Artifacts[0].Annotations["larch.snapshot.date"])

	snapshot := Snapshot{
		URL:       url,
		Title:     "",
		LibraryID: libraryID,
		Origin:    origin,
		ID:        snapshotID,
		Date:      date,
		Artifacts: make([]Artifact, 0),
	}

	for _, manifest := range index.Artifacts {
		artifact := Artifact{
			ContentType:     manifest.ContentType,
			LibraryID:       libraryID,
			ContentEncoding: manifest.ContentEncoding,
			Digest:          manifest.Digest,
			Size:            manifest.Size,
			Annotations:     manifest.Annotations,
		}

		snapshot.Artifacts = append(snapshot.Artifacts, artifact)

		blob, ok := i.blobs[artifact.Digest]
		if !ok {
			blob = Blob{
				ContentType:     manifest.ContentType,
				Libraries:       []string{},
				ContentEncoding: manifest.ContentEncoding,
				Digest:          manifest.Digest,
				Size:            manifest.Size,
			}
		}
		blob.Libraries = append(blob.Libraries, libraryID)
		i.blobs[artifact.Digest] = blob

		// Try to parse additional information from the Open Graph data
		if manifest.Annotations["larch.artifact.type"] == "vnd.larch.opengraph.meta.v1" {
			reader, err := snapshotReader.NextArtifactReader(ctx, manifest.Digest)
			if err == nil {
				var properties map[string][]string
				err := json.NewDecoder(reader).Decode(&properties)
				reader.Close()
				if err == nil {
					titles := properties["og:title"]
					if len(titles) > 0 {
						snapshot.Title = titles[0]
					}
				}
			}
		}
	}

	i.snapshots[origin+"/"+snapshotID] = snapshot
	return nil
}

// ListSnapshots implements Indexer.
func (i *InMemoryIndex) ListSnapshots(ctx context.Context, options *ListSnapshotsOptions) ([]Snapshot, error) {
	snapshots := make([]Snapshot, 0)
	for _, snapshot := range i.snapshots {
		if options != nil {
			if options.Origin != "" && snapshot.Origin != options.Origin {
				continue
			}
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// GetSnapshot implements Indexer.
func (i *InMemoryIndex) GetSnapshot(ctx context.Context, origin string, id string) (*Snapshot, error) {
	snapshot, ok := i.snapshots[origin+"/"+id]
	if !ok {
		return nil, ErrNotFound
	}

	return &snapshot, nil
}

func (i *InMemoryIndex) GetArtifact(ctx context.Context, origin string, id string, digest string) (*Artifact, error) {
	snapshot, ok := i.snapshots[origin+"/"+id]
	if !ok {
		return nil, ErrNotFound
	}

	var artifact *Artifact
	for _, a := range snapshot.Artifacts {
		if a.Digest == digest {
			artifact = &a
		}
	}
	if artifact == nil {
		return nil, ErrNotFound
	}

	return artifact, nil
}

func (i *InMemoryIndex) GetBlob(ctx context.Context, digest string) (*Blob, error) {
	blob, ok := i.blobs[digest]
	if !ok {
		return nil, ErrNotFound
	}

	return &blob, nil
}
