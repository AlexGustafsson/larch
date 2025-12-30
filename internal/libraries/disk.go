package libraries

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var _ LibraryWriter = (*DiskLibrary)(nil)
var _ LibraryReader = (*DiskLibrary)(nil)

type DiskLibrary struct {
	root *os.Root
}

func NewDiskLibrary(basePath string) (*DiskLibrary, error) {
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(basePath)
	if err != nil {
		return nil, err
	}

	return &DiskLibrary{
		root: root,
	}, nil
}

// GetOrigins implements LibraryReader.
func (d *DiskLibrary) GetOrigins(ctx context.Context) ([]string, error) {
	file, err := d.root.Open("snapshots")
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
func (d *DiskLibrary) GetSnapshots(ctx context.Context, origin string) ([]string, error) {
	file, err := d.root.Open(filepath.Join("snapshots", origin))
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
func (d *DiskLibrary) ReadSnapshot(ctx context.Context, id string) (SnapshotReader, error) {
	return newDiskSnapshotReader(d.root, id)
}

// WriteSnapshot implements LibraryWriter.
func (d *DiskLibrary) WriteSnapshot(ctx context.Context, id string) (SnapshotWriter, error) {
	return newDiskSnapshotWriter(d.root, id)
}

var _ SnapshotReader = (*diskSnapshotReader)(nil)

type diskSnapshotReader struct {
	root  *os.Root
	index SnapshotIndex
}

func newDiskSnapshotReader(root *os.Root, id string) (*diskSnapshotReader, error) {
	root, err := root.OpenRoot(filepath.Join("snapshots", id))
	if err != nil {
		return nil, err
	}

	indexFile, err := root.Open("index.json")
	if err != nil {
		root.Close()
		return nil, err
	}
	defer indexFile.Close()

	var index SnapshotIndex
	if err := json.NewDecoder(indexFile).Decode(&index); err != nil {
		return nil, err
	}

	return &diskSnapshotReader{
		root:  root,
		index: index,
	}, nil
}

// Index implements SnapshotReader.
func (d *diskSnapshotReader) Index() SnapshotIndex {
	return d.index
}

// NextReader implements SnapshotReader.
func (d *diskSnapshotReader) NextReader(ctx context.Context, name string) (DigestReadCloser, error) {
	file, err := d.root.Open(name)
	if err != nil {
		return nil, err
	}

	return newDigestReader(file), nil
}

// Close implements SnapshotReader.
func (d *diskSnapshotReader) Close() error {
	return d.root.Close()
}

type diskSnapshotWriter struct {
	root      *os.Root
	index     SnapshotIndex
	indexFile *os.File
}

func newDiskSnapshotWriter(root *os.Root, id string) (*diskSnapshotWriter, error) {
	if err := root.MkdirAll(filepath.Join("snapshots", id), 0755); err != nil {
		return nil, err
	}

	root, err := root.OpenRoot(filepath.Join("snapshots", id))
	if err != nil {
		return nil, err
	}

	created := false
	indexFile, err := root.OpenFile("index.json", os.O_RDWR, 0644)
	if errors.Is(err, os.ErrNotExist) {
		indexFile, err = root.OpenFile("index.json", os.O_CREATE|os.O_RDWR, 0644)
		created = true
	}
	if err != nil {
		root.Close()
		return nil, err
	}

	index := SnapshotIndex{
		MediaType: "application/vnd.larch.snapshot.index.v1+json",
		Manifests: make([]Manifest, 0),
	}
	if created {
		v, _ := json.MarshalIndent(&index, "", "  ")
		_, err := indexFile.Write(v)
		if err != nil {
			root.Close()
			return nil, err
		}
	} else {
		d, err := io.ReadAll(indexFile)
		if err != nil {
			root.Close()
			return nil, err
		}

		if err := json.Unmarshal(d, &index); err != nil {
			root.Close()
			return nil, err
		}
	}

	return &diskSnapshotWriter{
		root:      root,
		index:     index,
		indexFile: indexFile,
	}, nil
}

// NextWriter implements SnapshotWriter.
func (d *diskSnapshotWriter) NextWriter(ctx context.Context, path string) (DigestWriteCloser, error) {
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
func (d *diskSnapshotWriter) WriteFile(ctx context.Context, path string, data []byte) (int64, string, error) {
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

// WriteManifest implements SnapshotWriter.
func (d *diskSnapshotWriter) WriteManifest(ctx context.Context, manifest Manifest) error {
	_, err := d.indexFile.Seek(0, 0)
	if err != nil {
		return err
	}

	d.index.Manifests = append(d.index.Manifests, manifest)

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

func (d *diskSnapshotWriter) Close() error {
	return errors.Join(
		d.indexFile.Close(),
		d.root.Close(),
	)
}
