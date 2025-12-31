package disk

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.SnapshotReader = (*SnapshotReader)(nil)

type SnapshotReader struct {
	snapshotRoot *os.Root
	blobsRoot    *os.Root
	index        libraries.SnapshotIndex
}

func NewSnapshotReader(snapshotsRoot *os.Root, blobsRoot *os.Root, origin string, id string) (*SnapshotReader, error) {
	snapshotRoot, err := snapshotsRoot.OpenRoot(filepath.Join(origin, id))
	if err != nil {
		return nil, err
	}

	indexFile, err := snapshotRoot.Open("index.json")
	if err != nil {
		snapshotRoot.Close()
		return nil, err
	}
	defer indexFile.Close()

	var index libraries.SnapshotIndex
	if err := json.NewDecoder(indexFile).Decode(&index); err != nil {
		return nil, err
	}

	return &SnapshotReader{
		snapshotRoot: snapshotRoot,
		blobsRoot:    blobsRoot,
		index:        index,
	}, nil
}

// Index implements SnapshotReader.
func (s *SnapshotReader) Index() libraries.SnapshotIndex {
	return s.index
}

// NextArtifactReader implements SnapshotReader.
func (s *SnapshotReader) NextArtifactReader(ctx context.Context, digest string) (libraries.ArtifactReader, error) {
	return NewArtifactReader(s.blobsRoot, digest)
}

// Close implements SnapshotReader.
func (s *SnapshotReader) Close() error {
	return s.snapshotRoot.Close()
}
