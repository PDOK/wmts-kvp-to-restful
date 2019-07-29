package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"text/template"
)

var errorXmlTemplate = template.Must(
	template.New("errorXml").
		Funcs(template.FuncMap{"StringsJoin": strings.Join}).
		ParseFiles("errorXml.xml"))

func queryToPath(query map[string][]string) (path string, exception error) {
	// todo: we know these are the 6 expected queryparams rewrite to simpeler function.
	var layer, tilematrixset, tilematrix, tilecol, tilerow, format string

	var regex = regexp.MustCompile(`^.*:(.*)$`)

	for key, value := range query {
		value := value[0]

		if strings.ToLower(key) == "layer" {
			layer = value
		}

		if strings.ToLower(key) == "tilematrixset" {
			tilematrixset = value
		}

		if strings.ToLower(key) == "tilematrix" {
			groups := regex.FindAllStringSubmatch(value, -1)

			if groups != nil {
				tilematrix = groups[0][1]
			} else {
				tilematrix = value
			}
		}

		if strings.ToLower(key) == "tilerow" {
			tilerow = value
		}

		if strings.ToLower(key) == "tilecol" {
			tilecol = value
		}

		if strings.ToLower(key) == "format" {
			if value == "image/png8" {
				format = ".png"
			} else if value == "image/jpeg" {
				format = ".jpeg"
			} else {
				format = ".png"
			}
		}
	}

	path = "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + format

	return path, nil
}

func buildNewPath(urlPath, newQueryPath string) string {
	return strings.TrimRight(urlPath, "/") + newQueryPath
}

func validateTileQuery(query map[string][]string) []string {
	tileParams := [6]string{
		"layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format",
	}
	var missingParams []string

	for _, param := range tileParams {
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

// Operation is the constant type for the available WMTS Operations
type Operation string

// Const defining the available WMTS Operations
// Maybe check if al the required KVP are available
const (
	GetCapabilities Operation = "getcapabilities"
	GetTile         Operation = "gettile"
	GetFeatureInfo  Operation = "getfeatureinfo"
	None            Operation = "none"
)

// prio in order: GetCapabilities, GetTiles, GetFeatureInfo
func getOperation(query map[string][]string) Operation {
	var request string
	for key, values := range query {
		if strings.ToLower(key) == "request" {
			if len(values) > 1 {
				var countGetCapabilites, countGetTile, countGetFeatureInfo int

				for _, value := range values {

					switch strings.ToLower(value) {
					case string(GetCapabilities):
						countGetCapabilites = countGetCapabilites + 1
					case string(GetTile):
						countGetTile = countGetTile + 1
					case string(GetFeatureInfo):
						countGetFeatureInfo = countGetFeatureInfo + 1
					}
				}

				if countGetCapabilites > 0 {
					return GetCapabilities
				}

				if countGetTile > 0 {
					return GetTile
				}

				if countGetFeatureInfo > 0 {
					return GetFeatureInfo
				}

				return None
			}
			request = strings.ToLower(values[0])
		}
	}

	switch request {
	case string(GetTile):
		return GetTile
	case string(GetCapabilities):
		return GetCapabilities
	case string(GetFeatureInfo):
		return GetFeatureInfo
	default:
		return None
	}
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
		query := r.URL.Query()

		switch getOperation(query) {
		case GetTile:
			missingParams := validateTileQuery(query)
			var exception error = nil
			if len(missingParams) == 0 {
				var newPath string
				newPath, exception = queryToPath(query)
				r.URL.Path = buildNewPath(r.URL.Path, newPath)
				r.URL.RawQuery = ""
			} else if len(missingParams) < 6 {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
				exception = errorXmlTemplate.Execute(w, missingParams)
			}
			if exception != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.Write([]byte(`{"status": "rewrite went wrong"}`))
			}
		case GetCapabilities:
		case GetFeatureInfo:
		case None: // Probably a MissingParameterValue Error
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write([]byte(`{"status": "Not an valid WMTS KVP request"}`))
			return
		}

		proxy.ServeHTTP(w, r)
		return
	})

	log.Fatal(http.ListenAndServe(":9001", nil))
}
