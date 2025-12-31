package api

import (
	"encoding/json"
	"net/http"
	"slices"

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

	mux.HandleFunc("/api/v1/snapshots/{origin}/{id}/blobs/{algorithm}/{digest}", func(w http.ResponseWriter, r *http.Request) {
		// 	origin := r.PathValue("origin")
		// 	id := r.PathValue("id")
		// 	algorithm := r.PathValue("algorithm")
		// 	digest := r.PathValue("digest")

		// 	snapshotReader, err := library.ReadSnapshot(r.Context(), origin, id)
		// 	if err != nil {
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		// 	defer snapshotReader.Close()

		// 	index := snapshotReader.Index()

		// 	// TODO: Why not just merge manifests and layers on this level?
		// 	var m *libraries.Manifest
		// 	var l *libraries.Layer
		// loop:
		// 	for _, manifest := range index.Manifests {
		// 		for _, layer := range manifest.Layers {
		// 			if layer.Digest == algorithm+":"+digest {
		// 				m = &manifest
		// 				l = &layer
		// 				break loop
		// 			}
		// 		}
		// 	}

		// 	if m == nil || l == nil {
		// 		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		// 		return
		// 	}

		// 	reader, err := snapshotReader.NextReader(r.Context(), algorithm+":"+digest)
		// 	if err != nil {
		// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 		return
		// 	}
		// 	defer reader.Close()

		// 	w.Header().Set("Content-Type", l.MediaType)
		// 	// TODO: Content encoding, if Accept supports the content encoding, just
		// 	// provide it, otherwise we'll need to decompress, decrypt etc.
		// 	// TODO: Date, cache headers
		// 	io.Copy(w, reader)
	})

	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
