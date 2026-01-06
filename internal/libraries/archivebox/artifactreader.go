package archivebox

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.ArtifactReader = (*ArtifactReader)(nil)

type ArtifactReader struct {
	file   *os.File
	hash   hash.Hash
	reader io.Reader
}

func NewArtifactReader(root *os.Root, name string) (*ArtifactReader, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()

	return &ArtifactReader{
		file:   file,
		hash:   hash,
		reader: io.TeeReader(file, hash),
	}, nil
}

// Read implements libraries.DigestReadCloser.
func (a *ArtifactReader) Read(p []byte) (n int, err error) {
	return a.reader.Read(p)
}

// Close implements libraries.DigestReadCloser.
func (a *ArtifactReader) Close() error {
	return a.file.Close()
}

// Digest implements libraries.DigestReadCloser.
func (a *ArtifactReader) Digest() string {
	return "sha256:" + hex.EncodeToString(a.hash.Sum(nil))
}
