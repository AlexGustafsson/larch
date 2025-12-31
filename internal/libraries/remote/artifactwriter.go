package remote

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.ArtifactWriter = (*ArtifactWriter)(nil)

type ArtifactWriter struct {
	writer io.WriteCloser
	digest string
	errCh  chan error
	resCh  chan *http.Response
}

func NewArtifactWriter(client *http.Client, endpoint string, name string) (*ArtifactWriter, error) {
	reader, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, endpoint+"/api/v1/blobs", reader)
	if err != nil {
		return nil, err
	}

	artifactWriter := &ArtifactWriter{
		writer: writer,
		errCh:  make(chan error),
		resCh:  make(chan *http.Response),
	}

	go func() {
		res, err := client.Do(req)
		if err != nil {
			artifactWriter.errCh <- err
			close(artifactWriter.errCh)
			close(artifactWriter.resCh)
			return
		}

		artifactWriter.resCh <- res
		close(artifactWriter.errCh)
		close(artifactWriter.resCh)
	}()

	return artifactWriter, nil
}

// Write implements libraries.ArtifactWriter.
func (a *ArtifactWriter) Write(p []byte) (n int, err error) {
	return a.writer.Write(p)
}

// Close implements libraries.ArtifactWriter.
func (a *ArtifactWriter) Close() error {
	err := a.writer.Close()
	if err != nil {
		return err
	}

	select {
	case res := <-a.resCh:
		if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusConflict {
			return fmt.Errorf("got unexpected status: %d", res.StatusCode)
		}

		a.digest = res.Header.Get("X-Larch-Digest")
		return nil
	case err := <-a.errCh:
		return err
	}
}

// Digest implements libraries.ArtifactWriter.
func (a *ArtifactWriter) Digest() string {
	return a.digest
}
