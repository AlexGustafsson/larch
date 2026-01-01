package libraries

import (
	"context"
	"io"
)

type LibraryReader interface {
	// ReadSnapshot opens a [SnapshotReader] for the given origin and snapshot id.
	ReadSnapshot(context.Context, string, string) (SnapshotReader, error)
	// ReadArtifact opens a [ArtifactReader] for the artifact of the given
	// digest.
	ReadArtifact(context.Context, string) (ArtifactReader, error)
	// GetOrigins returns all origins of the library.
	GetOrigins(context.Context) ([]string, error)
	// GetSnapshots returns the id of all snapshots of an origin.
	GetSnapshots(context.Context, string) ([]string, error)
	// Close closes the library.
	Close() error
}

type LibraryWriter interface {
	// WriteSnapshot opens a [SnapshotWriter] for the given origin and snapshot
	// id.
	WriteSnapshot(context.Context, string, string) (SnapshotWriter, error)
}

type SnapshotReader interface {
	// Index returns the snapshot's index.
	Index() SnapshotIndex
	// NextArtifactReader returns a [ArtifactReader] for the given digest.
	NextArtifactReader(context.Context, string) (ArtifactReader, error)
	// Close closes the reader.
	Close() error
}

// TODO: It would be nice to optionally supply a digest, that would let the
// writer not calculate it and could allow for features like not writing the
// file to disk if it already exists or not even accepting it from a worker if
// it exists - saving us some disk writes. The Chrome archiver, for example, has
// all the data in-memory so it is trivial to hash before writing.
type SnapshotWriter interface {
	// NextArtifactWriter returns a [ArtifactWriter] for the given file name.
	// The name may be unused by the underlying implementation and should be
	// treated as a hint.
	NextArtifactWriter(context.Context, string) (ArtifactWriter, error)
	// WriteArtifact writes to a file by name and returns its size and digest.
	// The name may be unused by the underlying implementation and should be
	// treated as a hint.
	WriteArtifact(context.Context, string, []byte) (int64, string, error)
	// WriteArtifactManifest writes the manifest to the index.
	WriteArtifactManifest(context.Context, ArtifactManifest) error
	// Close closes the writer.
	Close() error
}

type ArtifactWriter interface {
	io.Writer
	io.Closer
	// Digest returns the written content's digest formatted like so:
	// <algorithm>:<hex-encoded digest>.
	// The value may be empty until the writer is closed.
	Digest() string
}

type ArtifactReader interface {
	io.Reader
	io.Closer
	// Digest returns the read content's digest formatted like so:
	// <algorithm>:<hex-encoded digest>.
	// The value may be empty until the reader is closed.
	Digest() string
}

// NOTE: There's no reason why an ArtifactManifest would not follow its parent's
// SnapshotIndex' content type.
type ArtifactManifest struct {
	ContentType string `json:"contentType"`
	Digest      string `json:"digest"`
	Size        int64  `json:"size"`
	// TODO: ContentEncoding for compression? Would map to annotation for OCI.
	// Use brotli for HTML, for example, leave that up to the archiver.
	ContentEncoding string            `json:"contentEncoding,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
}

type SnapshotIndex struct {
	// application/vnd.larch.snapshot.index.v1+json
	Schema    string             `json:"schema"`
	Artifacts []ArtifactManifest `json:"artifacts"`
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
