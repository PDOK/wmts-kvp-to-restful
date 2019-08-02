package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

// formats the WMTS query keys to lowercase, no WMTS query keys will be ignored
func formatQueryKeys(query url.Values) (url.Values, string, WMTSException) {
	// WMTSKeys to format, note: is only a union of the getcapabilities & gettile keys
	WMTSKeys := [9]string{"request", "service", "version", "layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}

	newWMTSQuery := url.Values{}
	noneWMTSQuery := "?"

	for key, values := range query {
		for _, wmtskey := range WMTSKeys {
			if wmtskey == strings.ToLower(key) {
				if len(values) != 1 {
					return nil, "", WMTSException{Error: fmt.Errorf("Multiple query values found for %s: %s", key, strings.Join(values, ",")), Code: "InvalidParameterValue", StatusCode: 400}
				}
				newWMTSQuery[wmtskey] = values
			}
		}
	}

	for key, values := range query {
		isWMTSkey := false
		for _, wmtskey := range WMTSKeys {
			if wmtskey == strings.ToLower(key) {
				isWMTSkey = true
				break
			}
		}
		if !isWMTSkey {
			for _, v := range values {
				noneWMTSQuery = noneWMTSQuery + fmt.Sprintf("%s=%s&", key, v)
			}
		}
	}
	return newWMTSQuery, strings.TrimRight(noneWMTSQuery, "&"), WMTSException{}
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

// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists/10510718
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		log.Fatal(err)
		return false
	}
	log.Println(err)
	return true
}

// TODO
// enable logging
// determine what to do with getfeatureinfo request...
// point those to 'default' end-point or ignore them...?
func main() {

	host := flag.String("host", "http://localhost", "Hostname to proxy with protocol, http/https and port")
	capabilitiestemplate := flag.String("t", "", "Optional GetCapabilities template file, if not set request will be proxied.")

	flag.Parse()

	if len(*host) == 0 {
		log.Fatal("No target host is configured")
		return
	}

	if !exists(*capabilitiestemplate) {
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
		WMTSquery, OtherQuery, err := formatQueryKeys(r.URL.Query())

		if err.Error != nil {
			sendError(err, w, r)
			return
		}

		if len(WMTSquery["request"]) == 1 {
			switch strings.ToLower(WMTSquery["request"][0]) {
			case "gettile":
				if procesGetTileRequest(WMTSquery, OtherQuery, w, r) {
					proxy.ServeHTTP(w, r)
				}
			case "getcapabilities":
				procesGetCapabilitiesRequest(WMTSquery, OtherQuery, *capabilitiestemplate, w, r)
			case "getfeatureinfo":
				//return GetFeatureInfo
			default:
				unknownRequest := WMTSException{Error: fmt.Errorf("Invalid request value: %s", WMTSquery["request"][0]), Code: "InvalidParameterValue", StatusCode: 400}
				sendError(unknownRequest, w, r)
			}
		} else {
			proxy.ServeHTTP(w, r)
		}
		return
	})

	log.Fatal(http.ListenAndServe(":9001", nil))
}
