package disk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type SnapshotWriter struct {
	snapshotRoot *os.Root
	blobsRoot    *os.Root
	index        libraries.SnapshotIndex
	indexFile    *os.File
}

func NewSnapshotWriter(snapshotsRoot *os.Root, blobsRoot *os.Root, origin string, id string) (*SnapshotWriter, error) {
	if err := snapshotsRoot.MkdirAll(filepath.Join(origin, id), 0755); err != nil {
		return nil, err
	}

	snapshotRoot, err := snapshotsRoot.OpenRoot(filepath.Join(origin, id))
	if err != nil {
		return nil, err
	}

	created := false
	indexFile, err := snapshotRoot.OpenFile("index.json", os.O_RDWR, 0644)
	if errors.Is(err, os.ErrNotExist) {
		indexFile, err = snapshotRoot.OpenFile("index.json", os.O_CREATE|os.O_RDWR, 0644)
		created = true
	}
	if err != nil {
		snapshotRoot.Close()
		blobsRoot.Close()
		return nil, err
	}

	index := libraries.SnapshotIndex{
		Schema:    "application/vnd.larch.snapshot.index.v1+json",
		Artifacts: make([]libraries.ArtifactManifest, 0),
	}
	if created {
		v, _ := json.MarshalIndent(&index, "", "  ")
		_, err := indexFile.Write(v)
		if err != nil {
			snapshotRoot.Close()
			blobsRoot.Close()
			return nil, err
		}
	} else {
		d, err := io.ReadAll(indexFile)
		if err != nil {
			snapshotRoot.Close()
			return nil, err
		}

		if err := json.Unmarshal(d, &index); err != nil {
			snapshotRoot.Close()
			return nil, err
		}
	}

	return &SnapshotWriter{
		snapshotRoot: snapshotRoot,
		blobsRoot:    blobsRoot,
		index:        index,
		indexFile:    indexFile,
	}, nil
}

// NextArtifactWriter implements SnapshotWriter.
func (d *SnapshotWriter) NextArtifactWriter(ctx context.Context, name string) (libraries.ArtifactWriter, error) {
	return NewArtifactWriter(d.snapshotRoot, d.blobsRoot, name)
}

// WriteArtifact implements SnapshotWriter.
func (d *SnapshotWriter) WriteArtifact(ctx context.Context, name string, data []byte) (int64, string, error) {
	w, err := d.NextArtifactWriter(ctx, name)
	if err != nil {
		return 0, "", err
	}
	defer w.Close()

	n, err := io.Copy(w, bytes.NewReader(data))
	if err != nil {
		return n, "", err
	}

	if err := w.Close(); err != nil {
		return n, "", err
	}

	digest := w.Digest()
	return n, digest, nil
}

// WriteArtifactManifest implements SnapshotWriter.
func (d *SnapshotWriter) WriteArtifactManifest(ctx context.Context, manifest libraries.ArtifactManifest) error {
	_, err := d.indexFile.Seek(0, 0)
	if err != nil {
		return err
	}

	d.index.Artifacts = append(d.index.Artifacts, manifest)

	_, err = d.indexFile.Seek(0, 0)
	if err != nil {
		return err
	}

	// Assume the file always grows, no need to truncate
	// TODO: Only write to index on close?
	encoder := json.NewEncoder(d.indexFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(d.index); err != nil {
		return err
	}

	return nil
}

func (d *SnapshotWriter) Close() error {
	return d.snapshotRoot.Close()
}
