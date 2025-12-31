package archivers

import (
	"context"

	urlpkg "net/url"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type ArchiveOrgArchiver struct {
}

func (a *ArchiveOrgArchiver) Archive(ctx context.Context, snapshotWriter libraries.SnapshotWriter, url string) error {
	// TODO: Actual client
	u, err := urlpkg.Parse(url)
	if err != nil {
		return err
	}

	u.Path = "web/" + u.Host + "/" + u.Path
	u.Host = "web.archive.org"
	u.Scheme = "https"
	u.RawQuery = ""

	size, digest, err := snapshotWriter.WriteFile(ctx, "archive.org/url.txt", []byte(u.String()))
	if err != nil {
		return err
	}

	return snapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
		Digest:      digest,
		ContentType: "text/plain",
		Size:        size,
		Annotations: map[string]string{
			"larch.artifact.path": "archive.org/url.txt",
			"larch.artifact.type": "vnd.larch.archive.org.url.v1",
		},
	})
}
