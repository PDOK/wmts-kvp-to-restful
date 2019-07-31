package main

import (
	"bytes"
	"net/http"
	"text/template"
)

// ExInvalidRequestValues is ExInvalidRequestValues
const ExInvalidRequestValues string = "Invalid number of request values"

// ExMultipleValuesFound is ExMultipleValuesFound
const ExMultipleValuesFound string = "Multiple query values found"

const errorXML = `<?xml version="1.0"?>
<ows:ExceptionReport xmlns:ows="http://www.opengis.net/ows/1.1"
                     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                     xsi:schemaLocation="http://www.opengis.net/ows/1.1 http://schemas.opengis.net/ows/1.1.0/owsExceptionReport.xsd"
                     version="1.0.0" xml:lang="en">
    <ows:Exception exceptionCode="gg">
        <ows:ExceptionText>ddd</ows:ExceptionText>
    </ows:Exception>
</ows:ExceptionReport>`

var errorXMLTemplate = template.Must(template.New("errorXML").Parse(errorXML))

// ExceptionCodeAndMessage is ExceptionCodeAndMessage
type ExceptionCodeAndMessage struct {
	Message string
}

func sendError(w http.ResponseWriter, r *http.Request) {
	errorMessage := ExceptionCodeAndMessage{Message: "sdsd"}
	buf := new(bytes.Buffer)
	errorXMLTemplate.Execute(buf, errorMessage)
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(buf.Bytes()))
	return
}
