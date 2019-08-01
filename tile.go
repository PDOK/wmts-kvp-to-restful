package main

import (
	"net/http"
	"net/url"
	"regexp"
)

func tileQueryToPath(query url.Values) (path string) {
	var regex = regexp.MustCompile(`^.*:(.*)$`)

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

	path = "/" +
		query["layer"][0] + "/" +
		query["tilematrixset"][0] + "/" +
		tilematrix + "/" +
		query["tilecol"][0] + "/" +
		query["tilerow"][0] +
		fileExtension

	return path
}

func procesGetTileRequest(query url.Values, w http.ResponseWriter, r *http.Request) {
	statusCode, path, contentType, _, _ := handleOperation(query, r, nil)
	if statusCode != 200 {
		w.WriteHeader(statusCode)
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if path != "" {
		r.URL.Path = path
		r.URL.RawQuery = ""
	}
}
