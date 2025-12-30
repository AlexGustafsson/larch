package libraries

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"io"
	"os"
	"path/filepath"
)

var _ LibraryWriter = (*BlobLibrary)(nil)

type BlobLibrary struct {
	BasePath string
}

// OpenSnapshot implements LibraryWriter.
func (d *BlobLibrary) OpenSnapshot(ctx context.Context, id string) (SnapshotWriter, error) {
	return newBlobWriter(d.BasePath, id)
}

var _ SnapshotWriter = (*blobWriter)(nil)

type blobWriter struct {
	root      *os.Root
	indexPath string
}

func newBlobWriter(basePath string, id string) (*blobWriter, error) {
	indexPath := filepath.Join("snapshots", id+".json")

	err := os.MkdirAll(filepath.Join(basePath, filepath.Dir(indexPath)), 0755)
	if err != nil {
		return nil, err
	}

	index := SnapshotManifest{
		MediaType: "application/vnd.larch.snapshot.index.v1+json",
		Manifests: []Manifest{},
	}

	root, err := os.OpenRoot(basePath)
	if err != nil {
		return nil, err
	}

	v, _ := json.MarshalIndent(&index, "", "  ")
	err = root.WriteFile(indexPath, v, 0644)
	if err != nil {
		return nil, err
	}

	return &blobWriter{
		root:      root,
		indexPath: indexPath,
	}, nil
}

// NextWriter implements SnapshotWriter.
func (d *blobWriter) NextWriter(ctx context.Context, path string) (DigestWriteCloser, error) {
	return newBlobFileWriter(d.root)
}

// WriteFile implements SnapshotWriter.
func (d *blobWriter) WriteFile(ctx context.Context, _path string, data []byte) (int64, string, error) {
	hash := sha256.New()
	hash.Write(data)

	digest := hex.EncodeToString(hash.Sum(nil))

	path := filepath.Join("blobs", "sha256", digest[:2], digest)

	err := d.root.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return 0, "", err
	}

	w, err := d.root.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, "", err
	}

	n, err := io.Copy(w, bytes.NewReader(data))
	if err != nil {
		return n, "", err
	}

	if err := w.Close(); err != nil {
		return n, "", err
	}

	return n, "sha256:" + digest, nil
}

// Index implements SnapshotWriter.
func (d *blobWriter) Index(ctx context.Context, manifest Manifest) error {
	// TODO: Atomic, or just keep in-memory?
	// TODO: Actually read from file
	// TODO: Actually append to manifests

	file, err := d.root.OpenFile(d.indexPath, os.O_RDWR, 0644)
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

var _ DigestWriteCloser = (*blobFileWriter)(nil)

type blobFileWriter struct {
	file   *os.File
	root   *os.Root
	hash   hash.Hash
	writer io.Writer
	digest string
}

func newBlobFileWriter(root *os.Root) (*blobFileWriter, error) {
	file, err := os.CreateTemp("", "larch-temp-*")
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	return &blobFileWriter{
		file:   file,
		root:   root,
		hash:   hash,
		writer: io.MultiWriter(file, hash),
	}, nil
}

func (w *blobFileWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *blobFileWriter) Close() error {
	w.digest = hex.EncodeToString(w.hash.Sum(nil))

	path := filepath.Join("blobs", "sha256", w.digest[:2], w.digest)

	err := w.root.MkdirAll(path, 0755)
	if err != nil {
		w.file.Close()
		return err
	}

	if _, err := w.file.Seek(0, 0); err != nil {
		w.file.Close()
		return err
	}

	writer, err := w.root.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		w.file.Close()
		return err
	}

	if err := w.file.Close(); err != nil {
		writer.Close()
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func (w *blobFileWriter) Digest() string {
	return "sha256:" + w.digest
}
