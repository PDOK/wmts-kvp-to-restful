package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

//GetCapabilitiesKeys is GetCapabilitiesKeys
var GetCapabilitiesKeys = [2]string{"request", "service"}

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

	if len(r.Header.Get("X-Forward-Host")) > 1 {
		host = r.Header.Get("X-Forward-Host")
	} else {
		host = r.URL.Host
	}

	if len(r.Header.Get("X-Script-Name")) > 1 {
		host = r.Header.Get("X-Script-Name")
	} else {
		host = r.URL.Path
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

func procesGetCapabilitiesRequest(query url.Values, otherquery string, template string, w http.ResponseWriter, r *http.Request) {
	missingParams := findMissingParams(query, GetCapabilitiesKeys[:])
	if len(missingParams) != 0 {
		err := WMTSException{Error: fmt.Errorf("Missing parameters: " + strings.Join(missingParams, ", ")), Code: "MissingParameterValue", StatusCode: 400}
		sendError(err, w, r)
		return
	}

	buf := new(bytes.Buffer)
	getCapabilitiesTemplate(template).Execute(buf, hostAndPath(r))
	w.Write([]byte(buf.Bytes()))
	w.Header().Set("Content-Type", "application/xml")
	return
}
