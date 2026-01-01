package worker

import (
	"context"
	"time"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.SnapshotWriter = (*Client)(nil)

type Client struct {
	Endpoint string
	Token    string
}

type JobRequest struct {
	Token string
	Job   Job
}

type Job struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	Requested time.Time `json:"requested"`
	Accepted  time.Time `json:"accepted,omitzero"`
	Started   time.Time `json:"started,omitzero"`
	Ended     time.Time `json:"ended,omitzero"`
	Error     string    `json:"error,omitempty"`
}

func (c *Client) GetJobRequests(ctx context.Context) (<-chan JobRequest, error) {
	panic("unimplemented")
}

func (c *Client) UpdateJob(ctx context.Context, job Job) error {
	panic("unimplemented")
}

// NextArtifactWriter implements libraries.SnapshotWriter.
func (c *Client) NextArtifactWriter(context.Context, string) (libraries.ArtifactWriter, error) {
	panic("unimplemented")
}

// WriteArtifact implements libraries.SnapshotWriter.
func (c *Client) WriteArtifact(context.Context, string, []byte) (int64, string, error) {
	panic("unimplemented")
}

// WriteArtifactManifest implements libraries.SnapshotWriter.
func (c *Client) WriteArtifactManifest(context.Context, libraries.ArtifactManifest) error {
	panic("unimplemented")
}

// Close implements libraries.SnapshotWriter.
func (c *Client) Close() error {
	panic("unimplemented")
}
