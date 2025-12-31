package disk

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.ArtifactWriter = (*ArtifactWriter)(nil)

type ArtifactWriter struct {
	name         string
	snapshotRoot *os.Root
	blobsRoot    *os.Root
	tempFile     *os.File
	hash         hash.Hash
	digest       string
	writer       io.Writer
}

func NewArtifactWriter(snapshotRoot *os.Root, blobsRoot *os.Root, name string) (*ArtifactWriter, error) {
	tempFile, err := os.CreateTemp("", "larch-temp-*")
	if err != nil {
		return nil, err
	}

	hash := sha256.New()

	return &ArtifactWriter{
		name:         name,
		snapshotRoot: snapshotRoot,
		blobsRoot:    blobsRoot,
		tempFile:     tempFile,
		hash:         hash,
		writer:       io.MultiWriter(tempFile, hash),
	}, nil
}

// Write implements libraries.DigestWriteCloser.
func (a *ArtifactWriter) Write(p []byte) (n int, err error) {
	return a.writer.Write(p)
}

// Close implements libraries.DigestWriteCloser.
func (a *ArtifactWriter) Close() error {
	_, err := a.tempFile.Seek(0, 0)
	if err != nil {
		_ = a.tempFile.Close()
		_ = os.Remove(a.tempFile.Name())
		return err
	}

	a.digest = string(hex.EncodeToString(a.hash.Sum(nil)))

	blobPath := filepath.Join("sha256", a.digest[0:2], a.digest[2:4], a.digest)

	err = a.blobsRoot.MkdirAll(filepath.Dir(blobPath), 0755)
	if err != nil {
		_ = a.tempFile.Close()
		_ = os.Remove(a.tempFile.Name())
		return err
	}

	file, err := a.blobsRoot.OpenFile(blobPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		_ = a.tempFile.Close()
		_ = os.Remove(a.tempFile.Name())
		return err
	}

	_, err = io.Copy(file, a.tempFile)
	_ = a.tempFile.Close()
	_ = os.Remove(a.tempFile.Name())
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	if err := a.snapshotRoot.MkdirAll(filepath.Dir(a.name), 0755); err != nil {
		return err
	}

	// NOTE: This can escape the root, but that's OK as the feature is meant for
	// humans / other tools - not larch itself. The symlinks are never used by us
	relativeBlobsDir, err := filepath.Rel(filepath.Join(a.snapshotRoot.Name(), filepath.Dir(a.name)), a.blobsRoot.Name())
	if err != nil {
		return err
	}

	err = a.snapshotRoot.Symlink(filepath.Join(relativeBlobsDir, blobPath), a.name)
	if err != nil {
		return err
	}

	return nil
}

// Digest implements libraries.DigestWriteCloser.
func (a *ArtifactWriter) Digest() string {
	return "sha256:" + a.digest
}
