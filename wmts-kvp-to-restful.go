package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

// Operation the constant type for the available WMTS Operations and an is ordered.
type Operation int

type OperationSlice []Operation

// These make OperationSlice sortable
func (o OperationSlice) Len() int           { return len(o) }
func (o OperationSlice) Less(i, j int) bool { return o[i] < o[j] }
func (o OperationSlice) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

var TileStrings = [6]string{"layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}

// Const defining the available WMTS Operations
// Maybe check if al the required KVP are available
const (
	GetCapabilities Operation = iota
	GetTile
	GetFeatureInfo
	None
)

// String representation of each operation
var operations = [...]string{
	"getcapabilities",
	"gettile",
	"getfeatureinfo",
	"none",
}

// String returns the English name of the operation.
func (o Operation) String() string { return operations[o] }

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

func operationFromStringSlice(a []string) OperationSlice {
	var result = make(OperationSlice, len(a))
	for i, v := range a {
		result[i] = operationFromString(v)
	}
	return result
}

var errorXmlTemplate = template.Must(
	template.New("errorXml").
		ParseFiles("errorXml.xml"))

var capabilitiesTemplate = template.Must(
	template.New("CapabilitiesXml").
		ParseFiles("WMTSCapabilities.xml"))

func lowerQueryKeys(query url.Values) url.Values {
	newQuery := url.Values{}
	for key, values := range query {
		newQuery[strings.ToLower(key)] = values
	}
	return newQuery
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
		fmt.Println(paramInQuery, param, len(missingParams))
	}
	return missingParams
}

// prio in order: GetCapabilities, GetTiles, GetFeatureInfo
func getOperation(query url.Values) Operation {
	request := query["request"]
	if request == nil {
		return None
	}
	requestTypes := operationFromStringSlice(request[:])
	sort.Sort(requestTypes)
	return requestTypes[0]
}

// TODO
// enable logging
// determine what to do with getcapabilities request and getfeatureinfo request...
// point those to 'default' end-point or ignore them...?
func main() {

	host := flag.String("host", "http://localhost", "Hostname to proxy with protocol, http/https and port")

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
		query := lowerQueryKeys(r.URL.Query())
		var exception error = nil
		switch getOperation(query) {
		case GetTile:
			log.Println("converting wmts tile request to kvp")
			missingParams := findMissingParams(query, TileStrings[:])
			if len(missingParams) == 0 {
				var newPath string
				newPath = tileQueryToPath(query)
				r.URL.Path = buildNewPath(r.URL.Path, newPath)
				r.URL.RawQuery = ""
			} else if len(missingParams) < 6 {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
				errorMessage := "missing parameters " + strings.Join(missingParams, ",")
				exception = errorXmlTemplate.Execute(w, errorMessage)
			}
		case GetCapabilities:
			log.Println("converting wmts getCapabilities request to kvp")
			w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
			// TODO: actually use the template syntax to fill in parameters.
			r.URL.Path = buildNewPath(r.URL.Path, "/v1_0/WMTSCapabilities.xml")
			r.URL.RawQuery = ""
			exception = capabilitiesTemplate.Execute(w, string(r.URL.RawPath))
		case GetFeatureInfo:
			exception = errors.New("not implemented")
		case None: // Probably a MissingParameterValue Error
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			exception = errorXmlTemplate.Execute(w, "Not an valid WMTS KVP request")
			log.Println("Invalid KVP request.")
			return
		}
		if exception != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			// TODO: this last possible error is unhandled:
			errorXmlTemplate.Execute(w, "rewrite went wrong")
			return
		}
		proxy.ServeHTTP(w, r)
		return
	})

	log.Fatal(http.ListenAndServe(":9001", nil))
}
