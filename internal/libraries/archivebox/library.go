package archivebox

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/AlexGustafsson/larch/internal/libraries"

	_ "modernc.org/sqlite"
)

var _ libraries.LibraryReader = (*Library)(nil)

type Library struct {
	root  *os.Root
	index *Index
}

func NewLibrary(basePath string, index *Index) (*Library, error) {
	root, err := os.OpenRoot(basePath)
	if err != nil {
		return nil, err
	}

	return &Library{
		root:  root,
		index: index,
	}, nil
}

// GetOrigins implements libraries.LibraryReader.
func (l *Library) GetOrigins(ctx context.Context) ([]string, error) {
	return slices.Collect(maps.Keys(l.index.Origins)), nil
}

// GetSnapshots implements libraries.LibraryReader.
func (l *Library) GetSnapshots(ctx context.Context, origin string) ([]string, error) {
	return l.index.SnapshotIDsByOrigin[origin], nil
}

// ReadArtifact implements libraries.LibraryReader.
func (l *Library) ReadArtifact(ctx context.Context, digest string) (libraries.ArtifactReader, error) {
	return NewArtifactReader(l.root, filepath.Join("archive", l.index.Blobs[digest]))
}

// ReadSnapshot implements libraries.LibraryReader.
func (l *Library) ReadSnapshot(ctx context.Context, origin string, id string) (libraries.SnapshotReader, error) {
	ids := l.index.SnapshotIDsByOrigin[origin]
	if !slices.Contains(ids, id) {
		return nil, os.ErrNotExist
	}

	return NewSnapshotReader(l.root, l.index, id)
}

// Close implements libraries.LibraryReader.
func (l *Library) Close() error {
	return l.root.Close()
}
