package operations

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

// Template used for the WMTS error responses
const errorXML = `<?xml version="1.0"?>
<ows:ExceptionReport xmlns:ows="http://www.opengis.net/ows/1.1"
                     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                     xsi:schemaLocation="http://www.opengis.net/ows/1.1 http://schemas.opengis.net/ows/1.1.0/owsExceptionReport.xsd"
                     version="1.0.0" xml:lang="en">
    <ows:Exception exceptionCode="{{ .Code }}">
        <ows:ExceptionText>{{ .Error }}</ows:ExceptionText>
    </ows:Exception>
</ows:ExceptionReport>`

// Exception interfact wraps three variables:
// Error
// Code
// Status
// Needed for WMTS error responses
type Exception interface {
	Error() string
	Code() string
	Status() int
}

// WMTSException grouping the error message variables together
type WMTSException struct {
	ErrorMessage string
	ErrorCode    string
	StatusCode   int
}

// Error returns available ErrorMessage
func (w WMTSException) Error() string {
	return w.ErrorMessage
}

// Code returns available ErrorCode
func (w WMTSException) Code() string {
	return w.ErrorCode
}

// Status returns available StatusCode
func (w WMTSException) Status() int {
	return w.StatusCode
}

// MissingParameterValue template
func MissingParameterValue(value string) Exception {
	return WMTSException{ErrorMessage: fmt.Sprintf("Missing parameter: %s", value),
		ErrorCode: "MissingParameterValue", StatusCode: 400}
}

// UnknownService template
func UnknownService(value string) Exception {
	return WMTSException{ErrorMessage: fmt.Sprintf("Missing SERVICE key or incorrect value, found: %s", value),
		ErrorCode: "MissingParameterValue", StatusCode: 400}
}

func InvalidParameterValue(parameter string, value string) Exception {
	return WMTSException{ErrorMessage: fmt.Sprintf("InvalidParameterValue for parameter: %s with value: %s",
		parameter, value), ErrorCode: "InvalidParameterValue", StatusCode: 400}
}

// SendError writes the error message to the response
func SendError(e Exception, w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	errorXMLTemplate := template.Must(template.New("errorXML").Parse(errorXML))
	errorXMLTemplate.Execute(buf, e)

	w.WriteHeader(e.Status())
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(buf.Bytes()))
}

// FindMissingParams compares the url.Values with the given keys
// and checks if any are missing
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
