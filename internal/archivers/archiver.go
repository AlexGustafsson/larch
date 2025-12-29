package archivers

import (
	"context"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type Archiver interface {
	Archive(context.Context, libraries.SnapshotWriter, string) error
}
