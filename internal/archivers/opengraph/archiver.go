package opengraph

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/AlexGustafsson/larch/internal/archivers"
	"github.com/AlexGustafsson/larch/internal/libraries"
)

var _ archivers.Archiver = (*Archiver)(nil)

type Archiver struct {
}

// Archive implements archivers.Archiver.
func (a *Archiver) Archive(ctx context.Context, snapshotWriter libraries.SnapshotWriter, url string) error {
	// In the happy path, where we want to read the document, it's cheaper to do
	// a single request for checking headers and reading the content. Prioritize
	// that over the case where we mistakenly request a non-HTML resource and drop
	// the response body where we could've just done a HEAD request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/html")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.Header.Get("Content-Type") != "text/html" && !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html;") {
		slog.Debug("Skipping OpenGraph as URL doesn't seem to point to a HTML document", slog.String("contentType", res.Header.Get("Content-Type")))
		res.Body.Close()
		return nil
	}

	decoder := xml.NewDecoder(res.Body)
	decoder.AutoClose = []string{"meta", "link"}
	decoder.Strict = false

	var document Document
	// Ignore any error, try to parse what we can
	_ = decoder.Decode(&document)
	res.Body.Close()

	properties := make(map[string][]string)
	for _, meta := range document.Meta {
		if !strings.HasPrefix(meta.Property, "og:") {
			continue
		}

		values, ok := properties[meta.Property]
		if !ok {
			values = make([]string, 0)
		}
		values = append(values, meta.Content)
		properties[meta.Property] = values
	}

	propertiesDocument, err := json.MarshalIndent(properties, "", "  ")
	if err != nil {
		return err
	}

	size, digest, err := snapshotWriter.WriteArtifact(ctx, "opengraph/meta.json", propertiesDocument)
	if err != nil {
		return err
	}

	err = snapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
		Digest:      digest,
		ContentType: "application/json",
		Size:        size,
		Annotations: map[string]string{
			"larch.artifact.path": "opengraph/meta.json",
			"larch.artifact.type": "vnd.larch.opengraph.meta.v1",
		},
	})
	if err != nil {
		return err
	}

	// Try to get the images
	for _, imageURL := range properties["og:image"] {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return err
		}

		extensions, _ := mime.ExtensionsByType(res.Header.Get("Content-Type"))

		extension := ""
		if len(extensions) > 0 {
			extension = extensions[0]
		}

		artifactWriter, err := snapshotWriter.NextArtifactWriter(ctx, "opengraph/image"+extension)
		if err != nil {
			res.Body.Close()
			return err
		}

		size, err := io.Copy(artifactWriter, res.Body)
		res.Body.Close()

		err = artifactWriter.Close()
		if err != nil {
			return err
		}

		digest := artifactWriter.Digest()

		err = snapshotWriter.WriteArtifactManifest(ctx, libraries.ArtifactManifest{
			Digest:      digest,
			ContentType: res.Header.Get("Content-Type"),
			Size:        size,
			Annotations: map[string]string{
				"larch.artifact.path": "opengraph/image" + extension,
				"larch.artifact.type": "vnd.larch.opengraph.image.v1",
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
