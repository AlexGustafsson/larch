package disk

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.LibraryWriter = (*Library)(nil)
var _ libraries.LibraryReader = (*Library)(nil)

type Library struct {
	snapshotsRoot *os.Root
	blobsRoot     *os.Root
}

func NewLibrary(basePath string) (*Library, error) {
	err := os.MkdirAll(filepath.Join(basePath, "snapshots"), 0755)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(basePath, "blobs"), 0755)
	if err != nil {
		return nil, err
	}

	snapshotsRoot, err := os.OpenRoot(filepath.Join(basePath, "snapshots"))
	if err != nil {
		return nil, err
	}

	blobsRoot, err := os.OpenRoot(filepath.Join(basePath, "blobs"))
	if err != nil {
		return nil, err
	}

	return &Library{
		snapshotsRoot: snapshotsRoot,
		blobsRoot:     blobsRoot,
	}, nil
}

// GetOrigins implements LibraryReader.
func (d *Library) GetOrigins(ctx context.Context) ([]string, error) {
	file, err := d.snapshotsRoot.Open(".")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries, err := file.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	origins := make([]string, len(entries))
	for i, entry := range entries {
		origins[i] = entry.Name()
	}

	return origins, nil
}

// GetSnapshots implements LibraryReader.
func (d *Library) GetSnapshots(ctx context.Context, origin string) ([]string, error) {
	file, err := d.snapshotsRoot.Open(origin)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries, err := file.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	snapshots := make([]string, len(entries))
	for i, entry := range entries {
		snapshots[i] = entry.Name()
	}

	return snapshots, nil
}

// ReadSnapshot implements LibraryReader.
func (d *Library) ReadSnapshot(ctx context.Context, origin string, id string) (libraries.SnapshotReader, error) {
	return NewSnapshotReader(d.snapshotsRoot, d.blobsRoot, origin, id)
}

// ReadArtifact implements LibraryReader.
func (d *Library) ReadArtifact(ctx context.Context, digest string) (libraries.ArtifactReader, error) {
	return NewArtifactReader(d.blobsRoot, digest)
}

// WriteSnapshot implements LibraryWriter.
func (d *Library) WriteSnapshot(ctx context.Context, origin string, id string) (libraries.SnapshotWriter, error) {
	return NewSnapshotWriter(d.snapshotsRoot, d.blobsRoot, origin, id)
}

func (l *Library) Close() error {
	return errors.Join(l.blobsRoot.Close(), l.snapshotsRoot.Close())
}
