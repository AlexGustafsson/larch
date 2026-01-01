package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type API struct {
	mux *http.ServeMux
}

func NewAPI(scheduler *Scheduler, library libraries.LibraryWriter) *API {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		workerID, requests, err := scheduler.RegisterWorker(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: Error handling, context deadline
		defer scheduler.UnregisterWorker(context.Background(), workerID)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		for request := range requests {
			event := "request"
			data, err := json.Marshal(request)
			if err != nil {
				return
			}

			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			<-time.After(2 * time.Second)
		}
	})

	mux.HandleFunc("PUT /api/v1/jobs/{job}", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Track job status
	})

	mux.HandleFunc("POST /api/v1/snapshots/{origin}/{id}/artifacts", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		id := r.PathValue("id")

		path := r.Header.Get("X-Larch-Path")
		// TODO: Check if file already exists, no need to send it or write it to
		// disk then
		// digest := r.Header.Get("X-Larch-Digest")

		// TODO: Support multiple libraries
		// TODO: Return conflict if already open?
		snapshotWriter, err := library.WriteSnapshot(r.Context(), origin, id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		writer, err := snapshotWriter.NextArtifactWriter(r.Context(), path)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		size, err := io.Copy(writer, r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = writer.Close()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		digest := writer.Digest()

		w.Header().Set("X-Larch-Size", strconv.FormatInt(size, 10))
		w.Header().Set("X-Larch-Digest", digest)
		w.WriteHeader(http.StatusCreated)
	})

	mux.HandleFunc("POST /api/v1/snapshots/{origin}/{id}/manifests", func(w http.ResponseWriter, r *http.Request) {
		origin := r.PathValue("origin")
		id := r.PathValue("id")

		var manifest libraries.ArtifactManifest
		if err := json.NewDecoder(r.Body).Decode(&manifest); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// TODO: Support multiple libraries
		// TODO: Return conflict if already open?
		snapshotWriter, err := library.WriteSnapshot(r.Context(), origin, id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = snapshotWriter.WriteArtifactManifest(r.Context(), manifest)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	return &API{
		mux: mux,
	}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
