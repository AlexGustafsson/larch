package archivebox

import (
	"context"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.SnapshotReader = (*SnapshotReader)(nil)

type SnapshotReader struct {
	id    string
	root  *os.Root
	index *Index
}

func NewSnapshotReader(root *os.Root, index *Index, id string) (*SnapshotReader, error) {
	root, err := root.OpenRoot(filepath.Join("archive", id))
	if err != nil {
		return nil, err
	}

	return &SnapshotReader{
		id:    id,
		root:  root,
		index: index,
	}, nil
}

// Index implements libraries.SnapshotReader.
func (s *SnapshotReader) Index() libraries.SnapshotIndex {
	return s.index.Snapshots[s.id]
}

// NextArtifactReader implements libraries.SnapshotReader.
func (s *SnapshotReader) NextArtifactReader(ctx context.Context, digest string) (libraries.ArtifactReader, error) {
	for _, artifact := range s.Index().Artifacts {
		if artifact.Digest == digest {
			return NewArtifactReader(s.root, artifact.Annotations["larch.artifact.path"])
		}
	}

	return nil, os.ErrNotExist
}

// Close implements libraries.SnapshotReader.
func (s *SnapshotReader) Close() error {
	return s.root.Close()
}
