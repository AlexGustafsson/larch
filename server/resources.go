package server

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ResourceHandler is an API for providing web resources for an WARC archive.
type ResourceHandler struct {
	router *mux.Router
	Server *Server
}

// NewResourceHandler creates a new resource handler.
func NewResourceHandler(server *Server, router *mux.Router) *ResourceHandler {
	resourceHandler := &ResourceHandler{
		router: router,
		Server: server,
	}

	router.HandleFunc("/{protocol}/{host}/{port}/", resourceHandler.getResource)
	router.HandleFunc("/{protocol}/{host}/{port}/{path}", resourceHandler.getResource)

	return resourceHandler
}

func (resourceHandler *ResourceHandler) getResource(response http.ResponseWriter, request *http.Request) {
	arguments := mux.Vars(request)

	port, err := strconv.ParseInt(arguments["port"], 10, 16)
	if err != nil {
		response.WriteHeader(400)
		fmt.Fprintf(response, "Bad Port")
		return
	}

	host := arguments["host"]
	if arguments["protocol"] == "https" {
		if port != 443 {
			host += ":" + strconv.FormatInt(port, 10)
		}
	} else if arguments["protocol"] == "http" {
		if port != 443 {
			host += ":" + strconv.FormatInt(port, 10)
		}
	} else {
		response.WriteHeader(400)
		fmt.Fprintf(response, "Unsupported Protocol")
		return
	}

	requestedURL := url.URL{Scheme: arguments["protocol"], Host: host, Path: arguments["path"]}

	record := resourceHandler.findWebRecord(requestedURL.String())
	if record == nil {
		response.WriteHeader(404)
		fmt.Fprintf(response, "Record Not Found")
		log.WithFields(log.Fields{"Type": "Web"}).Debugf("Record not found for URL '%s'", requestedURL.String())
		return
	}

	// Read the payload on demand if it's unavailable, but should be available
	// Note: there's currently no caching implemented. It's by design right
	// now to keep things simple. In the long run a time-based cache could
	// be relavant.
	payload := record.Payload
	if record.Header.ContentLength > 0 && payload == nil {
		if resourceHandler.Server.reader.Seekable {
			log.WithFields(log.Fields{"Type": "Web"}).Debugf("Payload for record %s is not loaded, reading", record.Header.RecordID)
			payload, err = resourceHandler.Server.reader.ReadPayload(record.Header)
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

func (resourceHandler *ResourceHandler) findWebRecord(requestedURL string) *warc.Record {
	for _, needle := range resourceHandler.Server.Archive.Records {
		if needle.Header.TargetURI == requestedURL && needle.Header.ContentType == "application/http;msgtype=response" {
			return needle
		}
	}

	// Match URLs with optional tailing slash
	if !strings.HasSuffix(requestedURL, "/") {
		return resourceHandler.findWebRecord(requestedURL + "/")
	}

	return nil
}
