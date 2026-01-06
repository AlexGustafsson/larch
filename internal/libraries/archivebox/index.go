package archivebox

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type Index struct {
	Origins             map[string]struct{}
	SnapshotIDsByOrigin map[string][]string
	Snapshots           map[string]libraries.SnapshotIndex
	Blobs               map[string]string
}

type Indexer struct {
	root *os.Root
}

func NewIndexer(basePath string) (*Indexer, error) {
	root, err := os.OpenRoot(basePath)
	if err != nil {
		return nil, err
	}

	return &Indexer{
		root: root,
	}, nil
}

func (i *Indexer) Index(ctx context.Context) (*Index, error) {
	index := &Index{
		Origins:             make(map[string]struct{}),
		SnapshotIDsByOrigin: make(map[string][]string),
		Snapshots:           make(map[string]libraries.SnapshotIndex),
		Blobs:               make(map[string]string),
	}

	snapshots, err := i.getSnapshots()
	if err != nil {
		return nil, err
	}

	for _, snapshot := range snapshots {
		snapshotURL, snapshotIndex, err := i.getSnapshot(snapshot)
		if err != nil {
			return nil, err
		}

		u, err := url.Parse(snapshotURL)
		if err != nil {
			return nil, err
		}

		index.Origins[u.Host] = struct{}{}

		ids, ok := index.SnapshotIDsByOrigin[u.Host]
		if !ok {
			ids = make([]string, 0)
		}
		ids = append(ids, snapshot)
		index.SnapshotIDsByOrigin[u.Host] = ids

		index.Snapshots[snapshot] = *snapshotIndex

		for _, artifact := range snapshotIndex.Artifacts {
			// NOTE: For now we don't care about duplicates here, just serve any blob
			path := filepath.Join(snapshot, artifact.Annotations["larch.artifact.path"])
			index.Blobs[artifact.Digest] = path
		}
	}

	return index, nil
}

func (i *Indexer) getSnapshots() ([]string, error) {
	file, err := i.root.Open("archive")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries, err := file.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	snapshots := make([]string, 0)
	for _, entry := range entries {
		snapshots = append(snapshots, entry.Name())
	}

	return snapshots, nil
}

func (i *Indexer) getSnapshot(id string) (string, *libraries.SnapshotIndex, error) {
	root, err := i.root.OpenRoot(filepath.Join("archive", id))
	if err != nil {
		return "", nil, err
	}

	indexFile, err := root.Open("index.json")
	if err != nil {
		return "", nil, err
	}
	defer indexFile.Close()

	var index struct {
		URL  string `json:"url"`
		Date string `json:"updated"`
	}
	if err := json.NewDecoder(indexFile).Decode(&index); err != nil {
		return "", nil, err
	}

	file, err := root.Open(".")
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	entries, err := file.ReadDir(-1)
	if err != nil {
		return "", nil, err
	}

	artifacts := make([]libraries.ArtifactManifest, 0)

	artifacts = append(artifacts, libraries.ArtifactManifest{
		ContentType: "application/vnd.larch.snapshot.manifest.v1+json",
		Digest:      "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Size:        0,
		Annotations: map[string]string{
			"larch.snapshot.url":  index.URL,
			"larch.snapshot.date": index.Date,
		},
	})

	for _, entry := range entries {
		// NOTE: For now we allow list known and supported file types
		var artifact *libraries.ArtifactManifest
		switch entry.Name() {
		case "archive.org.txt":
			artifact = &libraries.ArtifactManifest{
				ContentType: "text/plain",
				Annotations: map[string]string{
					"larch.artifact.path": "archive.org.txt",
					"larch.artifact.type": "vnd.archivebox.archive.org.url.v1",
				},
			}
		case "favicon.ico":
			artifact = &libraries.ArtifactManifest{
				ContentType: "image/x-icon",
				Annotations: map[string]string{
					"larch.artifact.path": "favicon.ico",
					"larch.artifact.type": "vnd.archivebox.favicon.v1",
				},
			}
		case "output.pdf":
			artifact = &libraries.ArtifactManifest{
				ContentType: "application/pdf",
				Annotations: map[string]string{
					"larch.artifact.path": "output.pdf",
					"larch.artifact.type": "vnd.archivebox.pdf.v1",
				},
			}
		case "screenshot.png":
			artifact = &libraries.ArtifactManifest{
				ContentType: "image/png",
				Annotations: map[string]string{
					"larch.artifact.path": "screenshot.png",
					"larch.artifact.type": "vnd.archivebox.screenshot.v1",
				},
			}
		case "singlefile.html":
			artifact = &libraries.ArtifactManifest{
				ContentType: "text/html",
				Annotations: map[string]string{
					"larch.artifact.path": "singlefile.html",
					"larch.artifact.type": "vnd.archivebox.singlefile.v1",
				},
			}
		}

		if artifact != nil {
			digest, size, err := sha256sum(root, entry.Name())
			if err != nil {
				return "", nil, err
			}
			artifact.Digest = digest
			artifact.Size = size
			artifacts = append(artifacts, *artifact)
		}
	}

	return index.URL, &libraries.SnapshotIndex{
		Schema:    "application/vnd.larch.snapshot.index.v1+json",
		Artifacts: artifacts,
	}, nil
}
