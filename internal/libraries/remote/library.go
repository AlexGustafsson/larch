package remote

import (
	"context"
	"net/http"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ libraries.LibraryWriter = (*Library)(nil)

type Library struct {
	Endpoint string
	Client   *http.Client
	// TODO: Propagate
	Token string
}

// WriteSnapshot implements libraries.LibraryWriter.
func (l *Library) WriteSnapshot(ctx context.Context, origin string, id string) (libraries.SnapshotWriter, error) {
	return &SnapshotWriter{
		Endpoint: l.Endpoint,
		Client:   l.Client,
	}, nil
}
