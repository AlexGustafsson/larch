package server

import (
	"fmt"
	"html"
	"net/http"
	"net/url"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ControlPanel is an API for providing operations on a WARC archive.
type ControlPanel struct {
	router *mux.Router
	Server *Server
}

// NewControlPanel creates a new control panel.
func NewControlPanel(server *Server, router *mux.Router) *ControlPanel {
	controlPanel := &ControlPanel{
		router: router,
		Server: server,
	}

	router.HandleFunc("", redirect("/larch/"))
	router.HandleFunc("/", controlPanel.listInterface)
	router.HandleFunc("/records", controlPanel.listRecords)
	router.HandleFunc("/record/{id}", controlPanel.getRecord)
	router.HandleFunc("/header/{id}", controlPanel.getHeader)
	router.HandleFunc("/payload/{id}", controlPanel.getPayload)
	router.HandleFunc("/sites", controlPanel.listSites)

	return controlPanel
}

func (controlPanel *ControlPanel) listInterface(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/html")
	response.WriteHeader(200)
	fmt.Fprint(response, "<html><body><ul><li><a href=\"/larch/records\">/larch/records</a> - List records</li><li><a href=\"/larch/sites\">/larch/sites</a> - List sites</li></ul></body></html>")
}

func (controlPanel *ControlPanel) listRecords(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/html")
	response.WriteHeader(200)
	fmt.Fprintln(response, "<html><body><table><thead><tr><th>ID</th><th>Content Type</th><th>Payload Size (Bytes)</th><th>Link To Record</th><th>Link To Header</th><th>Link To Payload</th></tr></thead><tbody>")
	for _, record := range controlPanel.Server.Archive.Records {
		fmt.Fprintf(response, "<tr><td>%s</td><td>%s</td><td>%d</td><td><a href=\"/larch/record/%s\">Record</a></td><td><a href=\"/larch/header/%s\">Header</a></td><td><a href=\"/larch/payload/%s\">Payload</a></td></tr>", html.EscapeString(record.Header.RecordID), record.Header.ContentType, record.Header.ContentLength, url.QueryEscape(record.Header.RecordID), url.QueryEscape(record.Header.RecordID), url.QueryEscape(record.Header.RecordID))
	}
	fmt.Fprintln(response, "</tbody></table></body></html>")
}

func (controlPanel *ControlPanel) listSites(response http.ResponseWriter, request *http.Request) {
	sites := make(map[string]bool)

	for _, record := range controlPanel.Server.Archive.Records {
		if record.Header.TargetURI != "" && record.Header.ContentType == "application/http;msgtype=request" {
			site, err := url.Parse(record.Header.TargetURI)
			if err == nil {
				sites[site.Host] = true
			}
		}
	}

	response.Header().Add("Content-Type", "text/html")
	response.WriteHeader(200)
	fmt.Fprintln(response, "<html><body><ul>")
	for site := range sites {
		fmt.Fprintf(response, "<li><p>%s</p></li>", html.EscapeString(site))
	}
	fmt.Fprintln(response, "</ul></body></html>")
}

func (controlPanel *ControlPanel) getRecord(response http.ResponseWriter, request *http.Request) {
	arguments := mux.Vars(request)
	id, err := url.QueryUnescape(arguments["id"])
	if err != nil {
		fmt.Fprintf(response, "Invalid record ID")
		return
	}

	record := controlPanel.findRecord(id)

	if record == nil {
		response.WriteHeader(404)
		fmt.Fprintf(response, "Record Not Found")
		return
	}

	// Read the payload on demand if it's unavailable, but should be available
	// Note: there's currently no caching implemented. It's by design right
	// now to keep things simple. In the long run a time-based cache could
	// be relavant.
	payload := record.Payload
	if record.Header.ContentLength > 0 && payload == nil {
		if controlPanel.Server.reader.Seekable {
			log.WithFields(log.Fields{"Type": "Web"}).Debugf("Payload for record %s is not loaded, reading", record.Header.RecordID)
			payload, err = controlPanel.Server.reader.ReadPayload(record.Header)
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

	response.Header().Add("Content-Type", "application/warc")
	response.WriteHeader(200)
	record.Header.Write(response)
	if payload != nil {
		payload.Write(response)
	}
}

func (controlPanel *ControlPanel) getHeader(response http.ResponseWriter, request *http.Request) {
	arguments := mux.Vars(request)
	id, err := url.QueryUnescape(arguments["id"])
	if err != nil {
		fmt.Fprintf(response, "Invalid record ID")
		return
	}

	record := controlPanel.findRecord(id)

	if record == nil {
		response.WriteHeader(404)
		fmt.Fprintf(response, "Record Not Found")
		return
	}

	response.Header().Add("Content-Type", "application/warc")
	response.WriteHeader(200)
	record.Header.Write(response)
}

func (controlPanel *ControlPanel) getPayload(response http.ResponseWriter, request *http.Request) {
	arguments := mux.Vars(request)
	id, err := url.QueryUnescape(arguments["id"])
	if err != nil {
		fmt.Fprintf(response, "Invalid record ID")
		return
	}

	record := controlPanel.findRecord(id)

	if record == nil {
		response.WriteHeader(404)
		fmt.Fprintf(response, "Record Not Found")
		return
	}

	// Read the payload on demand if it's unavailable, but should be available
	// Note: there's currently no caching implemented. It's by design right
	// now to keep things simple. In the long run a time-based cache could
	// be relavant.
	payload := record.Payload
	if record.Header.ContentLength > 0 && payload == nil {
		if controlPanel.Server.reader.Seekable {
			log.WithFields(log.Fields{"Type": "Web"}).Debugf("Payload for record %s is not loaded, reading", record.Header.RecordID)
			payload, err = controlPanel.Server.reader.ReadPayload(record.Header)
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

	response.Header().Add("Content-Type", record.Header.ContentType)
	response.WriteHeader(200)
	if payload != nil {
		payload.Write(response)
	}
}

func (controlPanel *ControlPanel) findRecord(id string) *warc.Record {
	for _, needle := range controlPanel.Server.Archive.Records {
		if needle.Header.RecordID == id {
			return needle
		}
	}

	return nil
}

func redirect(path string) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Location", path)
		response.WriteHeader(302)
	}
}
