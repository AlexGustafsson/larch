package worker

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/AlexGustafsson/larch/internal/libraries"
)

type API struct {
	mux *http.ServeMux
}

func NewAPI(scheduler *Scheduler, library libraries.LibraryWriter) *API {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Worker capabilities
		request, err := scheduler.GetJobRequest(r.Context(), nil)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(request)
	})

	mux.HandleFunc("PUT /api/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var job Job
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if job.ID != id {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// TODO: Handle context errors
		if err := scheduler.UpdateJob(r.Context(), job); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
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

		// TODO: Support multiple libraries (or library from strategy)
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
