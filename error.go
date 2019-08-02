package main

import (
	"bytes"
	"net/http"
	"text/template"
)

// MissingParameterValue 400 Bad request
// InvalidParameterValue 400 Bad request
// VersionNegotiationFailed 400 Bad request
// InvalidUpdateSequence 400 Bad request
// NoApplicableCode 500 Internal server error

// ExInvalidRequestValues is ExInvalidRequestValues
const ExInvalidRequestValues string = "Invalid number of request values"

const errorXML = `<?xml version="1.0"?>
<ows:ExceptionReport xmlns:ows="http://www.opengis.net/ows/1.1"
                     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                     xsi:schemaLocation="http://www.opengis.net/ows/1.1 http://schemas.opengis.net/ows/1.1.0/owsExceptionReport.xsd"
                     version="1.0.0" xml:lang="en">
    <ows:Exception exceptionCode="{{ .Code }}">
        <ows:ExceptionText>{{ .Error }}</ows:ExceptionText>
    </ows:Exception>
</ows:ExceptionReport>`

var errorXMLTemplate = template.Must(template.New("errorXML").Parse(errorXML))

// WMTSException is WMTSException
type WMTSException struct {
	Error      error
	Code       string
	StatusCode int
}

func sendError(e WMTSException, w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	errorXMLTemplate.Execute(buf, e)

	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(buf.Bytes()))
}
