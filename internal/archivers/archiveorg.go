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

	contentSize, contentDigest, err := snapshotWriter.WriteFile(ctx, "archive.org/url.txt", []byte(u.String()))
	if err != nil {
		return err
	}

	configSize, configDigest, err := snapshotWriter.WriteFile(ctx, "archive.org/config.json", []byte(`{}`))
	if err != nil {
		return err
	}

	return snapshotWriter.WriteManifest(ctx, libraries.Manifest{
		MediaType: "application/vnd.larch.artifact.manifest.v1+json",
		Config: libraries.Layer{
			Digest:    configDigest,
			MediaType: "vnd.larch.disk.config.v1+json",
			Size:      configSize,
			Annotations: map[string]string{
				"larch.artifact.path": "archive.org/config.json",
			},
		},
		Layers: []libraries.Layer{
			{
				Digest:    contentDigest,
				MediaType: "text/plain",
				Size:      contentSize,
				Annotations: map[string]string{
					"larch.artifact.path": "archive.org/url.txt",
				},
			},
		},
	})
}
