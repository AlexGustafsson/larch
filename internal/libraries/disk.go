package libraries

import (
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
		MediaType: "application/vnd.larch.snapshot.manifest.v1+json",
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
func (d *diskWriter) NextWriter(ctx context.Context, path string) (io.WriteCloser, error) {
	err := d.root.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	return d.root.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
}

// Index implements SnapshotWriter.
func (d *diskWriter) Index(ctx context.Context, manifest Manifest) error {
	// TODO: Atomic, or just keep in-memory?
	// TODO: Actually read from file
	// TODO: Actually append to manifests

	index := SnapshotManifest{
		MediaType: "application/vnd.larch.snapshot.manifest.v1+json",
		Manifests: []Manifest{
			manifest,
		},
	}

	v, _ := json.MarshalIndent(&index, "", "  ")
	return d.root.WriteFile("index.json", v, 0644)
}
