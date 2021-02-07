package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Server serves a WARC file locally.
type Server struct {
	reader          *warc.Reader
	Archive         *warc.File
	EnableInterface bool
}

// NewServer creates a new server for an archive.
func NewServer(reader *warc.Reader, archive *warc.File) *Server {
	return &Server{
		reader:          reader,
		Archive:         archive,
		EnableInterface: true,
	}
}

// Start starts the server on the current thread.
func (server *Server) Start(address string, port uint16) {
	router := mux.NewRouter()

	if server.EnableInterface {
		subrouter := router.PathPrefix("/larch").Subrouter()
		NewControlPanel(server, subrouter)
	}

	httpServer := &http.Server{
		Handler:      handlers.CompressHandler(handlers.CombinedLoggingHandler(os.Stdout, router)),
		Addr:         fmt.Sprintf("%s:%d", address, port),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.WithFields(log.Fields{"Type": "Web"}).Infof("Listening on TCP %v:%v", address, port)
	httpServer.ListenAndServe()
}
