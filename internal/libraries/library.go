package libraries

import (
	"context"
	"io"
)

type LibraryReader interface {
	OpenSnapshot(context.Context, string) (SnapshotReader, error)
	// TODO: Index
}

type LibraryWriter interface {
	OpenSnapshot(context.Context, string) (SnapshotWriter, error)
}

type SnapshotReader interface {
	NextReader(context.Context, string) (DigestReadCloser, error)
}

type SnapshotWriter interface {
	NextWriter(context.Context, string) (DigestWriteCloser, error)
	WriteFile(context.Context, string, []byte) (int64, string, error)
	Index(context.Context, Manifest) error
}

type DigestWriteCloser interface {
	io.Writer
	io.Closer
	Digest() string
}

type DigestReadCloser interface {
	io.Reader
	io.Closer
	Digest() string
}

type Manifest struct {
	MediaType string  `json:"mediaType"`
	Config    Layer   `json:"config,omitzero"`
	Layers    []Layer `json:"layers,omitempty"`
}

type Layer struct {
	Digest      string            `json:"digest"`
	MediaType   string            `json:"mediaType"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type SnapshotManifest struct {
	// application/vnd.larch.snapshot.manifest.v1+json
	MediaType string     `json:"mediaType"`
	Manifests []Manifest `json:"manifests"`
}

// IDEA: Index is just index, could be read on-boot? How would that work in
// registry? Get latest tag for all images, then image must be full url, not
// hostname.

// warc
// /libraries
// /libraries/disk
// /libraries/disk/index.sqlite
// /libraries/disk/snapshots/example.com/1231231.warc

// disk
// /libraries
// /libraries/disk
// /libraries/disk/index.sqlite
// /libraries/disk/snapshots/example.com/1231231/url.txt
// /libraries/disk/snapshots/example.com/12312312/index.json

// blob
// /libraries/blob/index.sqlite on-disk?
// /libraries/blob/snapshots/example.com/1721312312.json
// /libraries/blob/blobs/sha256/xxxxx
// /libraries/blob/blobs/sha256/xxxxx

// oci <=> blob on-disk? Why make any difference?
// tags: latest, shaid per snapshot etc. URL as name?
// registry.home.internal/larch/example-com/1231231231 <-manifest index, artifact
// registry.home.internal/blobs/sha256/xxxxx
// registry.home.internal/blobs/sha256/xxxxx
