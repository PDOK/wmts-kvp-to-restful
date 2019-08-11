package operations

import (
	"bytes"
	"net/http"
	"strconv"
	"text/template"
)

// HostAndPath is HostAndPath
type HostAndPath struct {
	Protocol string
	Host     string
	Path     string
}

func hostAndPath(r *http.Request) HostAndPath {
	var protocol, host, path string
	if len(r.Header.Get("X-Forwarded-Proto")) > 1 {
		protocol = r.Header.Get("X-Forwarded-Proto")
	} else {
		protocol = "http"
	}

	// Maybe something with port
	if len(r.Header.Get("X-Forward-Host")) > 1 {
		host = r.Header.Get("X-Forward-Host")
	} else if len(r.URL.Host) > 0 {
		host = r.URL.Host
	} else {
		host = "localhost:9001"
	}

	if len(r.Header.Get("X-Script-Name")) > 1 {
		path = r.Header.Get("X-Script-Name")
	} else {
		path = r.URL.Path
	}

	return HostAndPath{Protocol: protocol, Host: host, Path: path}
}

// GetCapabilitiesTemplate usage the path to return the template file
// and builds a template
func getCapabilitiesTemplate(path string) (*template.Template, Exception) {
	var capabilitiesTemplate = template.Must(template.ParseFiles(path))
	return capabilitiesTemplate, nil
}

// GetCapabilitiesKeys list of manitory WMTS getcapabilities key value pairs
func getCapabilitiesKeys() []string {
	return []string{"service", "request", "version"}
}

// ProcessGetCapabilitiesRequest if a template is given this will
// fill it in and writes it to the response
func ProcessGetCapabilitiesRequest(config *Config, w http.ResponseWriter, r *http.Request) Exception {
	buf := new(bytes.Buffer)
	t, _ := getCapabilitiesTemplate(config.Template)
	t.Execute(buf, hostAndPath(r))

	// Content-length header is needed for applications like QGIS
	// Maybe nicer way in calc capabilities documents size
	// For 'normal' size capabilities documents impact is low
	capabilities := buf.String()

	w.Header().Set("Server", "wmts-kvp-to-restful")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-length", strconv.Itoa(len(capabilities)))

	t.Execute(buf, hostAndPath(r))
	w.Write([]byte(capabilities))

	return nil
}
