package main

import (
	"log/slog"
	"net/http"
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

	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
