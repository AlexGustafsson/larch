package libraries

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

var _ LibraryWriter = (*DiskLibrary)(nil)

type DiskLibrary struct {
	BasePath string
}

// OpenSnapshot implements LibraryWriter.
func (d *DiskLibrary) OpenSnapshot(ctx context.Context, id string) (SnapshotWriter, error) {
	return newDiskWriter(d.BasePath, id)
}

var _ SnapshotWriter = (*diskWriter)(nil)

type diskWriter struct {
	root *os.Root
}

func newDiskWriter(basePath string, id string) (*diskWriter, error) {
	snapshootPath := filepath.Join(basePath, "snapshots", id)

	err := os.MkdirAll(snapshootPath, 0755)
	if err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(snapshootPath)
	if err != nil {
		return nil, err
	}

	index := SnapshotManifest{
		MediaType: "application/vnd.larch.snapshot.index.v1+json",
		Manifests: []Manifest{},
	}

	v, _ := json.MarshalIndent(&index, "", "  ")
	err = root.WriteFile("index.json", v, 0644)
	if err != nil {
		return nil, err
	}

	return &diskWriter{
		root: root,
	}, nil
}

// NextWriter implements SnapshotWriter.
func (d *diskWriter) NextWriter(ctx context.Context, path string) (DigestWriteCloser, error) {
	err := d.root.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	w, err := d.root.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return newDigestWriter(w), nil
}

// WriteFile implements SnapshotWriter.
func (d *diskWriter) WriteFile(ctx context.Context, path string, data []byte) (int64, string, error) {
	w, err := d.NextWriter(ctx, path)
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

// Index implements SnapshotWriter.
func (d *diskWriter) Index(ctx context.Context, manifest Manifest) error {
	// TODO: Atomic, or just keep in-memory?
	// TODO: Actually read from file
	// TODO: Actually append to manifests

	file, err := d.root.OpenFile("index.json", os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var index SnapshotManifest
	if err := json.NewDecoder(file).Decode(&index); err != nil {
		return err
	}

	index.Manifests = append(index.Manifests, manifest)

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Assume the file always grows, no need to truncate
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(index); err != nil {
		return err
	}

	return nil
}
