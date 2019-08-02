package error

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
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

// SendError is SendError
func SendError(e WMTSException, w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	errorXMLTemplate.Execute(buf, e)

	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(buf.Bytes()))
}

// FindMissingParams is FindMissingParams
func FindMissingParams(query url.Values, queryParams []string) []string {
	var missingParams []string
	for _, param := range queryParams {

		paramInQuery := false
		for key := range query {
			paramInQuery = paramInQuery || (strings.ToLower(key) == param)
		}
		if !paramInQuery {
			missingParams = append(missingParams, param)
		}
	}
	return missingParams
}
