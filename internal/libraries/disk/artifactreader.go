package disk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.ArtifactReader = (*ArtifactReader)(nil)

type ArtifactReader struct {
	file   *os.File
	hash   hash.Hash
	reader io.Reader
}

func NewArtifactReader(blobsRoot *os.Root, digest string) (*ArtifactReader, error) {
	algorithm, digest, ok := strings.Cut(digest, ":")
	if !ok {
		return nil, fmt.Errorf("invalid digest")
	}

	file, err := blobsRoot.Open(filepath.Join(algorithm, digest[0:2], digest[2:4], digest))
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
