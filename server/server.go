package server

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Site            string
}

// NewServer creates a new server for an archive.
func NewServer(reader *warc.Reader, archive *warc.File, site string) *Server {
	return &Server{
		reader:          reader,
		Archive:         archive,
		EnableInterface: true,
		Site:            site,
	}
}

// Start starts the server on the current thread.
func (server *Server) Start(address string, port uint16) {
	router := mux.NewRouter()

	if server.EnableInterface {
		subrouter := router.PathPrefix("/larch").Subrouter()
		NewControlPanel(server, subrouter)
	}

	// Handle any path
	router.PathPrefix("/").HandlerFunc(server.serve)

	httpServer := &http.Server{
		Handler:      handlers.CompressHandler(handlers.CombinedLoggingHandler(os.Stdout, router)),
		Addr:         fmt.Sprintf("%s:%d", address, port),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.WithFields(log.Fields{"Type": "Web"}).Infof("Listening on TCP %v:%v", address, port)
	httpServer.ListenAndServe()
}

func (server *Server) serve(response http.ResponseWriter, request *http.Request) {
	for _, record := range server.Archive.Records {
		if record.Header.TargetURI != "" {
			targetURL, err := url.Parse(record.Header.TargetURI)
			if err != nil {
				continue
			}

			matches := targetURL.Host == server.Site && targetURL.Path == request.URL.Path
			if !matches {
				continue
			}

			// If the record matches but is not a HTTP response, look for another request
			// as there may be more than one record per URI
			if record.Header.ContentType != "application/http;msgtype=response" {
				log.WithFields(log.Fields{"Type": "Web"}).Debugf("Skipping record for site '%v' that was not a HTTP response", targetURL)
				continue
			}

			// Read the payload on demand if it's unavailable, but should be available
			// Note: there's currently no caching implemented. It's by design right
			// now to keep things simple. In the long run a time-based cache could
			// be relavant.
			payload := record.Payload
			if record.Header.ContentLength > 0 && payload == nil {
				if server.reader.Seekable {
					log.WithFields(log.Fields{"Type": "Web"}).Debugf("Payload for record %s is not loaded, reading", record.Header.RecordID)
					payload, err = server.reader.ReadPayload(record.Header)
					if err != nil {
						response.WriteHeader(500)
						fmt.Fprintf(response, "Unable To Read Payload")
						log.WithFields(log.Fields{"Type": "Web"}).Error(err)
						return
					}
				} else {
					response.WriteHeader(503)
					fmt.Fprintf(response, "Payload Not Loaded")
					return
				}
			}

			if payload == nil {
				response.WriteHeader(500)
				fmt.Fprintf(response, "Got Empty Payload")
				return
			}

			payloadReader := bufio.NewReader(payload.Reader())
			httpResponse, err := http.ReadResponse(payloadReader, nil)
			if err != nil {
				response.WriteHeader(500)
				log.WithFields(log.Fields{"Type": "Web"}).Error(err)
				return
			}

			// Remove automatically added headers
			response.Header().Del("Content-Type")
			response.Header().Del("Vary")
			response.Header().Del("Date")
			response.Header().Del("Transfer-Encoding")

			// Add back the original headers
			for key, values := range httpResponse.Header {
				if len(values) >= 1 {
					response.Header().Set(key, values[0])
					for _, value := range values[1:] {
						response.Header().Add(key, value)
					}
				}
			}

			// Write the original response and status code
			response.WriteHeader(httpResponse.StatusCode)
			io.Copy(response, httpResponse.Body)
			return
		}
	}

	response.WriteHeader(404)
	fmt.Fprintf(response, "Not Found")
}
