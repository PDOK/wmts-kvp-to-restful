package operations

import (
	"bytes"
	"net/http"
	"net/url"
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
	} else {
		host = "localhost"
	}

	if len(r.Header.Get("X-Script-Name")) > 1 {
		path = r.Header.Get("X-Script-Name")
	} else {
		path = r.URL.Path
	}

	return HostAndPath{
		Protocol: protocol,
		Host:     host,
		Path:     path}
}

func getCapabilitiesTemplate(path string) *template.Template {
	var capabilitiesTemplate = template.Must(template.ParseFiles(path))
	return capabilitiesTemplate
}

// GetCapabilitiesKeys is public
func GetCapabilitiesKeys() []string {
	return []string{"request", "service"}
}

// ProcesGetCapabilitiesRequest is public
func ProcesGetCapabilitiesRequest(query url.Values, template string, w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	getCapabilitiesTemplate(template).Execute(buf, hostAndPath(r))
	w.Write([]byte(buf.Bytes()))
	w.Header().Set("Content-Type", "application/xml")
	return
}
