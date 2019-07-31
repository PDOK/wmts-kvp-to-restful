package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"text/template"
)

// Operation the constant type for the available WMTS Operations and an is ordered.
type Operation string

var TileStrings = [6]string{"layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}

// Const defining the available WMTS Operations
// Maybe check if al the required KVP are available
const (
	GetCapabilities Operation = "getcapabilities"
	GetTile         Operation = "gettile"
	GetFeatureInfo  Operation = "getfeatureinfo"
	None            Operation = "none"
)

func operationFromString(s string) Operation {
	switch strings.ToLower(s) {
	case "getcapabilities":
		return GetCapabilities
	case "gettile":
		return GetTile
	case "getfeatureinfo":
		return GetFeatureInfo
	default:
		return None
	}
}

var errorXmlTemplate = template.Must(
	template.New("errorXml.xml").
		Parse(errorXml))

func formatQuery(query url.Values) (url.Values, error) {
	newQuery := url.Values{}
	for key, values := range query {
		if len(values) != 1 {
			if key != "sections" {
				return nil, errors.New(ExMultipleValuesFound)
			}
		}
		newQuery[strings.ToLower(key)] = values
	}
	return newQuery, nil
}

func getCapabilitiesTemplate(path string) *template.Template {
	var capabilitiesTemplate = template.Must(
		template.New("WMTSCapabilities.xml").
			ParseFiles(path))
	return capabilitiesTemplate
}

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

func buildNewPath(urlPath, newQueryPath string) string {
	return strings.TrimRight(urlPath, "/") + newQueryPath
}

func findMissingParams(query url.Values, queryParams []string) []string {
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

// ExInvalidRequestValues is ExInvalidRequestValues
const ExInvalidRequestValues string = "Invalid number of request values"

// ExMultipleValuesFound is ExMultipleValuesFound
const ExMultipleValuesFound string = "Multiple query values found"

// prio in order: GetCapabilities, GetTiles, GetFeatureInfo
func getOperation(query url.Values) (operation Operation, exception error) {
	request := query["request"]
	if len(request) != 1 {
		return None, errors.New(ExInvalidRequestValues)
	}
	return operationFromString(request[0]), nil
}

func handleOperation(query url.Values, r *http.Request, incomingException error) (
	statusCode int, path string, contentType string, operation Operation, exception error) {
	if incomingException != nil {
		statusCode = http.StatusBadRequest
		contentType = "application/xml; charset=UTF-8"
		exception = incomingException
	} else {
		operation, exception = getOperation(query)
		if exception != nil {
			statusCode = http.StatusBadRequest
			contentType = "application/xml; charset=UTF-8"
			return statusCode, path, contentType, operation, exception
		}
		switch operation {
		case GetTile:
			log.Println("Converting wmts tile request to kvp.")
			missingParams := findMissingParams(query, TileStrings[:])
			if len(missingParams) == 0 {
				statusCode = http.StatusOK
				path = buildNewPath(r.URL.Path, tileQueryToPath(query))
			} else {
				statusCode = http.StatusBadRequest
				contentType = "application/xml; charset=UTF-8"
				exception = errors.New("Missing parameters: '" + strings.Join(missingParams, "', '") + "'.")
			}
		case GetCapabilities:
			log.Println("Converting wmts getCapabilities request to kvp.")
			statusCode = http.StatusOK
			contentType = "application/xml; charset=UTF-8"
			path = buildNewPath(r.URL.Path, "/v1_0/WMTSCapabilities.xml")
		case GetFeatureInfo:
			statusCode = http.StatusInternalServerError
			contentType = "application/xml; charset=UTF-8"
			exception = errors.New("GetFeatureInfo not implemented.")
		case None: // Probably a MissingParameterValue Error
			statusCode = http.StatusInternalServerError
			contentType = "application/xml; charset=UTF-8"
			exception = errors.New("Not an valid WMTS KVP request.")
		}
	}
	return statusCode, path, contentType, operation, exception
}

// TODO
// enable logging
// determine what to do with getfeatureinfo request...
// point those to 'default' end-point or ignore them...?
func main() {

	host := flag.String("host", "http://localhost", "Hostname to proxy with protocol, http/https and port")
	capabilitiesfile := flag.String("capabilitiesfile", "WMTSCapabilities.xml", "Optional GetCapabilities template file, if not set request will be proxied.")

	flag.Parse()

	if len(*host) == 0 {
		log.Fatal("No target host is configured")
		return
	}

	origin, _ := url.Parse(*host)

	director := func(req *http.Request) {
		req.URL.Host = origin.Host
		req.URL.Scheme = origin.Scheme
		req.Host = origin.Host
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
	}

	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(`{"health": "OK"}`))
		return
	})

	log.Println("wmts-kvp-to-restful started")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var xmlparseException error

		query, formatException := formatQuery(r.URL.Query())
		statusCode, path, contentType, operation, exception := handleOperation(query, r, formatException)
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

		if exception != nil {
			xmlparseException = errorXmlTemplate.Execute(w, exception.Error())
		} else if operation == GetCapabilities && *capabilitiesfile != "" {
			xmlparseException = getCapabilitiesTemplate(*capabilitiesfile).Execute(w, path)
		}

		if xmlparseException != nil {
			log.Fatal(xmlparseException.Error())
		}

		proxy.ServeHTTP(w, r)
		return
	})

	log.Fatal(http.ListenAndServe(":9001", nil))
}
