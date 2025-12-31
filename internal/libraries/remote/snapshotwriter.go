package remote

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.SnapshotWriter = (*SnapshotWriter)(nil)

type SnapshotWriter struct {
	Endpoint string
	Client   *http.Client
}

// NextArtifactWriter implements libraries.SnapshotWriter.
func (s *SnapshotWriter) NextArtifactWriter(ctx context.Context, name string) (libraries.ArtifactWriter, error) {
	return NewArtifactWriter(s.Client, s.Endpoint, name)
}

// WriteArtifact implements libraries.SnapshotWriter.
func (s *SnapshotWriter) WriteArtifact(ctx context.Context, name string, data []byte) (int64, string, error) {
	w, err := s.NextArtifactWriter(ctx, name)
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

// WriteArtifactManifest implements libraries.SnapshotWriter.
func (s *SnapshotWriter) WriteArtifactManifest(context.Context, libraries.ArtifactManifest) error {
	panic("unimplemented")
}

// Close implements libraries.SnapshotWriter.
func (s *SnapshotWriter) Close() error {
	panic("unimplemented")
}
