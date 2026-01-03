package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(index indexers.Indexer, library libraries.LibraryReader) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		snapshots, err := index.ListSnapshots(r.Context(), nil)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		embeddedSnapshots := make([]Snapshot, 0)

		for _, snapshot := range snapshots {
			embeddedArtifacts := make([]Artifact, 0)
			for _, artifact := range snapshot.Artifacts {
				algorithm, digest, _ := strings.Cut(artifact.Digest, ":")
				embeddedArtifacts = append(embeddedArtifacts, Artifact{
					ContentType:     artifact.ContentType,
					ContentEncoding: artifact.ContentEncoding,
					Digest:          artifact.Digest,
					Size:            artifact.Size,
					Links: ArtifactLinks{
						Curies: []Link{
							{
								Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
								Name:      "larch",
								Templated: true,
							},
						},
						Self: Link{
							Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts/%s/%s", snapshot.Origin, snapshot.ID, algorithm, digest),
						},
						Snapshot: Link{
							Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
						},
						Origin: Link{
							Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
						},
						Blob: Link{
							Href: fmt.Sprintf("/api/v1/blobs/%s/%s", algorithm, digest),
						},
					},
				})
			}

			embeddedSnapshots = append(embeddedSnapshots, Snapshot{
				ID:     snapshot.ID,
				URL:    snapshot.URL,
				Origin: snapshot.Origin,
				Date:   snapshot.Date,
				Embedded: SnapshotEmbedded{
					Artifacts: embeddedArtifacts,
				},
				Links: SnapshotLinks{
					Curies: []Link{
						{
							Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
							Name:      "larch",
							Templated: true,
						},
					},
					Self: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
					},
					Origin: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
					},
					Artifacts: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts", snapshot.Origin, snapshot.ID),
					},
				},
			})
		}

		page := Page[SnapshotPageEmbedded]{
			Page:  1,
			Size:  30,
			Count: len(snapshots),
			Total: len(snapshots),
			Embedded: SnapshotPageEmbedded{
				Snapshots: embeddedSnapshots,
			},
			Links: PageLinks{
				Curies: []Link{
					{
						Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
						Name:      "larch",
						Templated: true,
					},
				},
				Self: Link{
					Href: "/api/v1/snapshots?page=1",
				},
				First: Link{
					Href: "/api/v1/snapshots?page=1",
				},
				Last: Link{
					Href: "/api/v1/snapshots?page=1",
				},
				Page: Link{
					Href:      "/api/v1/snapshots?page={page}",
					Templated: true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	})

	mux.HandleFunc("GET /api/v1/snapshots/{origin}", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")

		snapshots, err := index.ListSnapshots(r.Context(), &indexers.ListSnapshotsOptions{Origin: origin})
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		snapshots = slices.DeleteFunc(snapshots, func(snapshot indexers.Snapshot) bool {
			return snapshot.Origin != origin
		})

		// TODO: Shared formatting logic with snapshot endpoint
		embeddedSnapshots := make([]Snapshot, 0)

		for _, snapshot := range snapshots {
			embeddedArtifacts := make([]Artifact, 0)
			for _, artifact := range snapshot.Artifacts {
				algorithm, digest, _ := strings.Cut(artifact.Digest, ":")
				embeddedArtifacts = append(embeddedArtifacts, Artifact{
					ContentType:     artifact.ContentType,
					ContentEncoding: artifact.ContentEncoding,
					Digest:          artifact.Digest,
					Size:            artifact.Size,
					Links: ArtifactLinks{
						Curies: []Link{
							{
								Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
								Name:      "larch",
								Templated: true,
							},
						},
						Self: Link{
							Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts/%s/%s", snapshot.Origin, snapshot.ID, algorithm, digest),
						},
						Snapshot: Link{
							Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
						},
						Origin: Link{
							Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
						},
						Blob: Link{
							Href: fmt.Sprintf("/api/v1/blobs/%s/%s", algorithm, digest),
						},
					},
				})
			}

			embeddedSnapshots = append(embeddedSnapshots, Snapshot{
				ID:     snapshot.ID,
				URL:    snapshot.URL,
				Origin: snapshot.Origin,
				Date:   snapshot.Date,
				Embedded: SnapshotEmbedded{
					Artifacts: embeddedArtifacts,
				},
				Links: SnapshotLinks{
					Curies: []Link{
						{
							Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
							Name:      "larch",
							Templated: true,
						},
					},
					Self: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
					},
					Origin: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
					},
					Artifacts: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts", snapshot.Origin, snapshot.ID),
					},
				},
			})
		}

		page := Page[SnapshotPageEmbedded]{
			Page:  1,
			Size:  30,
			Count: len(snapshots),
			Total: len(snapshots),
			Embedded: SnapshotPageEmbedded{
				Snapshots: embeddedSnapshots,
			},
			Links: PageLinks{
				Curies: []Link{
					{
						Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
						Name:      "larch",
						Templated: true,
					},
				},
				Self: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s?page=1", origin),
				},
				First: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s?page=1", origin),
				},
				Last: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s?page=1", origin),
				},
				Page: Link{
					Href:      fmt.Sprintf("/api/v1/snapshots/%s?page={page}", origin),
					Templated: true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	})

	mux.HandleFunc("GET /api/v1/snapshots/{origin}/{id}", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		id := r.PathValue("id")

		snapshot, err := index.GetSnapshot(r.Context(), origin, id)
		if err == indexers.ErrNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: Shared formatting logic with snapshot endpoint
		embeddedArtifacts := make([]Artifact, 0)
		for _, artifact := range snapshot.Artifacts {
			algorithm, digest, _ := strings.Cut(artifact.Digest, ":")
			embeddedArtifacts = append(embeddedArtifacts, Artifact{
				ContentType:     artifact.ContentType,
				ContentEncoding: artifact.ContentEncoding,
				Digest:          artifact.Digest,
				Size:            artifact.Size,
				Links: ArtifactLinks{
					Curies: []Link{
						{
							Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
							Name:      "larch",
							Templated: true,
						},
					},
					Self: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts/%s/%s", snapshot.Origin, snapshot.ID, algorithm, digest),
					},
					Snapshot: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
					},
					Origin: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
					},
					Blob: Link{
						Href: fmt.Sprintf("/api/v1/blobs/%s/%s", algorithm, digest),
					},
				},
			})
		}

		res := Snapshot{
			ID:     snapshot.ID,
			URL:    snapshot.URL,
			Origin: snapshot.Origin,
			Date:   snapshot.Date,
			Embedded: SnapshotEmbedded{
				Artifacts: embeddedArtifacts,
			},
			Links: SnapshotLinks{
				Curies: []Link{
					{
						Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
						Name:      "larch",
						Templated: true,
					},
				},
				Self: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
				},
				Origin: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
				},
				Artifacts: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts", snapshot.Origin, snapshot.ID),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("GET /api/v1/snapshots/{origin}/{id}/artifacts", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		id := r.PathValue("id")

		snapshot, err := index.GetSnapshot(r.Context(), origin, id)
		if err == indexers.ErrNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: Shared formatting logic with snapshot endpoint
		embeddedArtifacts := make([]Artifact, 0)
		for _, artifact := range snapshot.Artifacts {
			algorithm, digest, _ := strings.Cut(artifact.Digest, ":")
			embeddedArtifacts = append(embeddedArtifacts, Artifact{
				ContentType:     artifact.ContentType,
				ContentEncoding: artifact.ContentEncoding,
				Digest:          artifact.Digest,
				Size:            artifact.Size,
				Links: ArtifactLinks{
					Curies: []Link{
						{
							Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
							Name:      "larch",
							Templated: true,
						},
					},
					Self: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts/%s/%s", snapshot.Origin, snapshot.ID, algorithm, digest),
					},
					Snapshot: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
					},
					Origin: Link{
						Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
					},
					Blob: Link{
						Href: fmt.Sprintf("/api/v1/blobs/%s/%s", algorithm, digest),
					},
				},
			})
		}

		// NOTE: The page format is only used to adhere to the rest of the list
		// types, but listing artifacts of a snapshot is not really a use case given
		// how few there will be
		page := Page[ArtifactPageEmbedded]{
			Page:  1,
			Size:  30,
			Count: len(embeddedArtifacts),
			Total: len(embeddedArtifacts),
			Embedded: ArtifactPageEmbedded{
				Artifacts: embeddedArtifacts,
			},
			Links: PageLinks{
				Curies: []Link{
					{
						Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
						Name:      "larch",
						Templated: true,
					},
				},
				Self: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts?page=1", origin, id),
				},
				First: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts?page=1", origin, id),
				},
				Last: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts?page=1", origin, id),
				},
				Page: Link{
					Href:      fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts?page={page}", origin, id),
					Templated: true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	})

	mux.HandleFunc("GET /api/v1/snapshots/{origin}/{id}/artifacts/{algorithm}/{digest}", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		id := r.PathValue("id")
		algorithm := r.PathValue("algorithm")
		digest := r.PathValue("digest")

		snapshot, err := index.GetSnapshot(r.Context(), origin, id)
		if err == indexers.ErrNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		artifact, err := index.GetArtifact(r.Context(), algorithm+":"+digest)
		if err == indexers.ErrNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: Shared formatting logic with snapshot endpoint
		res := Artifact{
			ContentType:     artifact.ContentType,
			ContentEncoding: artifact.ContentEncoding,
			Digest:          artifact.Digest,
			Size:            artifact.Size,
			Links: ArtifactLinks{
				Curies: []Link{
					{
						Href:      "https://github.com/AlexGustafsson/larch/blob/main/docs/api.md#{rel}",
						Name:      "larch",
						Templated: true,
					},
				},
				Self: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s/artifacts/%s/%s", snapshot.Origin, snapshot.ID, algorithm, digest),
				},
				Snapshot: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s/%s", snapshot.Origin, snapshot.ID),
				},
				Origin: Link{
					Href: fmt.Sprintf("/api/v1/snapshots/%s", snapshot.Origin),
				},
				Blob: Link{
					Href: fmt.Sprintf("/api/v1/blobs/%s/%s", algorithm, digest),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("HEAD /api/v1/blobs/{algorithm}/{digest}", func(w http.ResponseWriter, r *http.Request) {
		digest := r.PathValue("algorithm") + ":" + r.PathValue("digest")

		// Special case for the well-known empty blob
		if digest == "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		artifact, err := index.GetArtifact(r.Context(), digest)
		if err == indexers.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get artifact", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", artifact.ContentType)
		w.Header().Set("Content-Length", strconv.FormatInt(artifact.Size, 10))
		// TODO: Content encoding, if Accept supports the content encoding, just
		// provide it, otherwise we'll need to decompress, decrypt etc.
		// TODO: Date, cache headers
		// TODO: Content-Digest?
	})

	mux.HandleFunc("GET /api/v1/blobs/{algorithm}/{digest}", func(w http.ResponseWriter, r *http.Request) {
		digest := r.PathValue("algorithm") + ":" + r.PathValue("digest")

		// Special case for the well-known empty blob
		if digest == "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		artifact, err := index.GetArtifact(r.Context(), digest)
		if err == indexers.ErrNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get artifact", slog.Any("error", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: Support multiple libraries
		artifactReader, err := library.ReadArtifact(r.Context(), digest)
		if err != nil {
			slog.Error("Failed to read artifact", slog.Any("error", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", artifact.ContentType)
		w.Header().Set("Content-Length", strconv.FormatInt(artifact.Size, 10))
		// TODO: Content encoding, if Accept supports the content encoding, just
		// provide it, otherwise we'll need to decompress, decrypt etc.
		// TODO: Date, cache headers
		// TODO: Content-Digest?
		if _, err := io.Copy(w, artifactReader); err != nil {
			artifactReader.Close()
			return
		}

		if err := artifactReader.Close(); err != nil {
			return
		}

		actualDigest := artifactReader.Digest()
		if artifact.Digest != actualDigest {
			slog.Warn("Recorded artifact digest does not match actual digest", slog.String("expected", digest), slog.String("actual", actualDigest))
		}
	})

	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
