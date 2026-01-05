package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/template"

	"github.com/AlexGustafsson/larch/internal/api"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(client *api.Client) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFS(templates, "templates/index.html.gotmpl")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		snapshots, err := client.GetSnapshots(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, snapshots); err != nil {
			slog.Error("Failed to render template", slog.Any("error", err))
			// Fallthrough
		}
	})

	mux.HandleFunc("GET /snapshots/{origin}/{snapshot}", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		snapshotID := r.PathValue("snapshot")

		t, err := template.ParseFS(templates, "templates/snapshot.html.gotmpl")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		snapshot, err := client.GetSnapshot(r.Context(), origin, snapshotID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, snapshot); err != nil {
			slog.Error("Failed to render template", slog.Any("error", err))
			// Fallthrough
		}
	})

	mux.HandleFunc("GET /snapshots/{origin}/{snapshot}/artifacts/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		snapshotID := r.PathValue("snapshot")

		path := strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/snapshots/%s/%s/artifacts/", origin, snapshotID))

		snapshot, err := client.GetSnapshot(r.Context(), origin, snapshotID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var artifact *api.Artifact
		for _, a := range snapshot.Embedded.Artifacts {
			if a.Annotations["larch.artifact.path"] == path {
				artifact = &a
			}
		}
		if artifact == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		_ = client.CopyBlob(r.Context(), w, artifact.Digest)
	})

	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
