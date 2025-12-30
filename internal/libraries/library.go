package libraries

import (
	"context"
	"io"
)

type LibraryReader interface {
	ReadSnapshot(context.Context, string) (SnapshotReader, error)
	GetOrigins(context.Context) ([]string, error)
	GetSnapshots(context.Context, string) ([]string, error)
}

type LibraryWriter interface {
	WriteSnapshot(context.Context, string) (SnapshotWriter, error)
}

type SnapshotReader interface {
	Index() SnapshotIndex
	NextReader(context.Context, string) (DigestReadCloser, error)
	Close() error
}

type SnapshotWriter interface {
	NextWriter(context.Context, string) (DigestWriteCloser, error)
	WriteFile(context.Context, string, []byte) (int64, string, error)
	WriteManifest(context.Context, Manifest) error
	Close() error
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
	Digest string `json:"digest"`
	// TODO: ContentEncoding for compression? Would map to annotation for OCI.
	// Use brotli for HTML, for example, leave that up to the archiver.
	MediaType   string            `json:"mediaType"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type SnapshotIndex struct {
	// application/vnd.larch.snapshot.index.v1+json
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
// registry.home.internal/larch/index:example.com <- what to put here?
// registry.home.internal/larch/example.com:1231231231 <-manifest index, artifact
// registry.home.internal/blobs/sha256/xxxxx
// registry.home.internal/blobs/sha256/xxxxx
