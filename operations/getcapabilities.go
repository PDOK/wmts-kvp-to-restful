package operations

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

// HostAndPath is HostAndPath
type HostAndPath struct {
	Protocol string
	Host     string
	Path     string
}

func hostAndPath(r *http.Request) HostAndPath {
	protocol := "http"
	host := r.Host
	path := r.URL.Path

	xfp, ok := r.Header["X-Forwarded-Proto"]
	if ok {

		protocol = xfp[0]
	}

	xfh, ok := r.Header["X-Forwarded-Host"]
	if ok {
		// When multiple proxy enviroments are past, like some corporate infrastructures.
		// Headers can be rewritten or appended, when a comma separated list is used it
		// will take the `first` entry from the header
		groups := strings.Split(xfh[0], ",")
		host = groups[0]
	}

	// Used by K8s proxy
	xfu, ok := r.Header["X-Forwarded-Uri"]
	if ok {
		path = xfu[0]
	}

	// Used by Traefik on PathPrefixStrip rules
	xfpr, ok := r.Header["X-Forwarded-Prefix"]
	if ok {
		path = xfpr[0]
	}

	retval := HostAndPath{Protocol: protocol, Host: host, Path: path}

	fmt.Printf("%+v\n", retval)

	return retval
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
