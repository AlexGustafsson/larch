package archivers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type ArchiveOrgArchiver struct {
}

func (a *ArchiveOrgArchiver) Archive(ctx context.Context, snapshotWriter libraries.SnapshotWriter, url string) error {
	w, err := snapshotWriter.NextWriter(ctx, "archive.org/url.txt")
	if err != nil {
		return err
	}
	defer w.Close()

	fileHash := sha256.New()

	contents := bytes.NewReader([]byte("https://web.archive.org/web/github.com/pocket-id/pocket-id"))
	fileSize, err := io.Copy(w, io.TeeReader(contents, fileHash))
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	w, err = snapshotWriter.NextWriter(ctx, "archive.org/config.json")
	if err != nil {
		return err
	}
	defer w.Close()

	configFileHash := sha256.New()

	contents = bytes.NewReader([]byte(`{}`))
	configFileSize, err := io.Copy(w, io.TeeReader(contents, configFileHash))
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return snapshotWriter.Index(ctx, libraries.Manifest{
		MediaType: "application/vnd.larch.artifact.manifest.v1+json",
		Config: libraries.Layer{
			Digest:    "sha256:" + hex.EncodeToString(configFileHash.Sum(nil)),
			MediaType: "vnd.larch.disk.config.v1+json",
			Size:      configFileSize,
			Annotations: map[string]string{
				"larch.artifact.path": "archive.org/config.json",
			},
		},
		Layers: []libraries.Layer{
			{
				Digest:    "sha256:" + hex.EncodeToString(fileHash.Sum(nil)),
				MediaType: "text/plain",
				Size:      fileSize,
				Annotations: map[string]string{
					"larch.artifact.path": "archive.org/url.txt",
				},
			},
		},
	})
}
