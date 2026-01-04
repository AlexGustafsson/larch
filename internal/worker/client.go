package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type Client struct {
	Endpoint string
	Client   *http.Client
}

func (c *Client) GetJobRequest(ctx context.Context) (*JobRequest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/jobs", c.Endpoint), nil)
	if err != nil {
		return nil, err
	}

	// TODO: E-Tag?

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var jobRequest JobRequest
	if err := json.NewDecoder(res.Body).Decode(&jobRequest); err != nil {
		return nil, err
	}

	return &jobRequest, nil
}

var _ libraries.SnapshotWriter = (*JobClient)(nil)

type JobClient struct {
	Origin     string
	SnapshotID string
	Endpoint   string
	Token      string
	Client     *http.Client
}

func (c *JobClient) UpdateJob(ctx context.Context, job Job) error {
	body, err := json.Marshal(&job)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/api/v1/jobs/%s", c.Endpoint, url.PathEscape(job.ID)), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	// TODO: E-Tag?

	res, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

// NextArtifactWriter implements libraries.SnapshotWriter.
func (c *JobClient) NextArtifactWriter(ctx context.Context, name string) (libraries.ArtifactWriter, error) {
	slog.Debug("Requesting artifact writer")
	reader, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/snapshots/%s/%s/artifacts", c.Endpoint, c.Origin, c.SnapshotID), reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	req.Header.Set("X-Larch-Name", name)

	errCh := make(chan error)
	digestCh := make(chan string)
	go func() {
		defer close(errCh)
		defer close(digestCh)

		res, err := c.Client.Do(req)
		if err != nil {
			errCh <- err
			return
		}

		// TODO: Handle conflict etc.
		if res.StatusCode != http.StatusCreated {
			errCh <- fmt.Errorf("unexpected status code: %d", res.StatusCode)
			return
		}

		digest := res.Header.Get("X-Larch-Digest")
		digestCh <- digest
	}()

	return &artifactWriter{
		errCh:    errCh,
		digestCh: digestCh,
		writer:   writer,
	}, nil
}

// WriteArtifact implements libraries.SnapshotWriter.
func (c *JobClient) WriteArtifact(ctx context.Context, name string, data []byte) (int64, string, error) {
	slog.Debug("Writing artifact")
	w, err := c.NextArtifactWriter(ctx, name)
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
func (c *JobClient) WriteArtifactManifest(ctx context.Context, manifest libraries.ArtifactManifest) error {
	slog.Debug("Writing artifact manifest")
	body, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/snapshots/%s/%s/manifests", c.Endpoint, c.Origin, c.SnapshotID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	res, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

// Close implements libraries.SnapshotWriter.
func (c *JobClient) Close() error {
	return nil
}

var _ libraries.ArtifactWriter = (*artifactWriter)(nil)

type artifactWriter struct {
	errCh    <-chan error
	digestCh <-chan string
	writer   io.WriteCloser
	digest   string
}

// Write implements libraries.ArtifactWriter.
func (a *artifactWriter) Write(p []byte) (n int, err error) {
	return a.writer.Write(p)
}

// Close implements libraries.ArtifactWriter.
func (a *artifactWriter) Close() error {
	if err := a.writer.Close(); err != nil {
		return err
	}

	select {
	case err := <-a.errCh:
		return err
	case digest := <-a.digestCh:
		a.digest = digest
		return nil
	}
}

// Digest implements libraries.ArtifactWriter.
func (a *artifactWriter) Digest() string {
	return a.digest
}
