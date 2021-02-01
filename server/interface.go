package server

import (
	"fmt"
	"html"
	"net/http"
	"net/url"

	"github.com/AlexGustafsson/larch/formats/warc"
	"github.com/gorilla/mux"
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

	router.HandleFunc("/", redirect("/larch/list"))
	router.HandleFunc("", redirect("/larch/"))
	router.HandleFunc("/list", controlPanel.listArchive)
	router.HandleFunc("/record/{id}", controlPanel.getRecord)
	router.HandleFunc("/header/{id}", controlPanel.getHeader)
	router.HandleFunc("/payload/{id}", controlPanel.getPayload)

	return controlPanel
}

func (controlPanel *ControlPanel) listArchive(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/html")
	response.WriteHeader(200)
	fmt.Fprintln(response, "<html><body><table><thead><tr><th>ID</th><th>Content Type</th><th>Payload Size (Bytes)</th><th>Link To Record</th><th>Link To Header</th><th>Link To Payload</th></tr></thead><tbody>")
	for _, record := range controlPanel.Server.Archive.Records {
		fmt.Fprintf(response, "<tr><td>%s</td><td>%s</td><td>%d</td><td><a href=\"/larch/record/%s\">Record</a></td><td><a href=\"/larch/header/%s\">Header</a></td><td><a href=\"/larch/payload/%s\">Payload</a></td></tr>", html.EscapeString(record.Header.RecordID), record.Header.ContentType, record.Header.ContentLength, url.QueryEscape(record.Header.RecordID), url.QueryEscape(record.Header.RecordID), url.QueryEscape(record.Header.RecordID))
	}
	fmt.Fprintln(response, "</tbody></table></body></html>")
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

	record.Write(response)
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

	if record.Payload != nil {
		record.Payload.Write(response)
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
