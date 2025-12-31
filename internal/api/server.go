package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/AlexGustafsson/larch/internal/indexers"
	"github.com/AlexGustafsson/larch/internal/libraries"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(index indexers.Indexer, library libraries.LibraryReader) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		snapshots, err := index.ListSnapshots(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshots)
	})

	mux.HandleFunc("/api/v1/snapshots/{origin}", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")

		// TODO: Other API?
		snapshots, err := index.ListSnapshots(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		snapshots = slices.DeleteFunc(snapshots, func(snapshot indexers.Snapshot) bool {
			return snapshot.Origin != origin
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshots)
	})

	mux.HandleFunc("/api/v1/snapshots/{origin}/{id}", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		id := r.PathValue("id")

		// TODO: Other API?
		snapshots, err := index.ListSnapshots(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var snapshot *indexers.Snapshot
		for _, s := range snapshots {
			if s.Origin == origin && s.ID == id {
				snapshot = &s
				break
			}
		}

		if snapshot == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})

	mux.HandleFunc("/api/v1/blobs/{algorithm}/{digest}", func(w http.ResponseWriter, r *http.Request) {
		digest := r.PathValue("algorithm") + ":" + r.PathValue("digest")

		artifact, err := index.GetArtifact(r.Context(), digest)
		if err == indexers.ErrNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get artifact", slog.Any("error", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

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
