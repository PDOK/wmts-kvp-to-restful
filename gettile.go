package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
)

// GetTileKeys containing the manditory query keys
var GetTileKeys = [8]string{"request", "service", "layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}

var regex = regexp.MustCompile(`^.*:(.*)$`)

const restTemplate = `/{{ .Layer }}/{{ .Tilematrixset }}/{{ .Tilematrix }}/{{ .Tilecol }}/{{ .Tilerow }}{{ .Fileextension }}`

// RestParameters is RestParameters
type RestParameters struct {
	Layer         string
	Tilematrixset string
	Tilematrix    string
	Tilecol       string
	Tilerow       string
	Fileextension string
}

func tileQueryToPath(query url.Values) string {
	tilematrix := query["tilematrix"][0]
	groups := regex.FindAllStringSubmatch(tilematrix, -1)
	if groups != nil {
		tilematrix = groups[0][1]
	}

	var fileExtension string
	switch query["format"][0] {
	case "image/png8":
		fileExtension = ".png"
	case "image/jpeg":
		fileExtension = ".jpeg"
	default:
		fileExtension = ".png"
	}

	restParameters := &RestParameters{Layer: query["layer"][0], Tilematrixset: query["tilematrixset"][0], Tilematrix: tilematrix, Tilecol: query["tilecol"][0], Tilerow: query["tilerow"][0], Fileextension: fileExtension}

	buf := new(bytes.Buffer)
	template.Must(template.New("restTemplate").Parse(restTemplate)).Execute(buf, restParameters)

	return buf.String()
}

func procesGetTileRequest(query url.Values, otherquery string, w http.ResponseWriter, r *http.Request) bool {
	missingParams := findMissingParams(query, GetTileKeys[:])
	if len(missingParams) != 0 {
		err := WMTSException{Error: fmt.Errorf("Missing parameters: " + strings.Join(missingParams, ", ")), Code: "MissingParameterValue", StatusCode: 400}
		sendError(err, w, r)

		return false
	}
	r.URL.Path = strings.TrimRight(r.URL.Path, "/") + tileQueryToPath(query)
	if otherquery != "" {
		r.URL.RawQuery = otherquery
	}
	return true
}
