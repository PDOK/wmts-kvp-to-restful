package operations

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Config used for storing application startup parameters
type Config struct {
	Host     string
	Template string
	Logging  bool
}

// Convert all the keys to lowercase and checks if there is only
// one value for every key
func keysToLowerAndFilter(query url.Values) (url.Values, Exception) {
	newquery := url.Values{}

	for key, values := range query {
		if len(values) != 1 {
			return nil, WMTSException{ErrorMessage: fmt.Sprintf("Multiple query values found for %s: %s", key, strings.Join(values, ",")), ErrorCode: "InvalidParameterValue", StatusCode: 400}
		}
		newquery[strings.ToLower(key)] = values
	}
	return newquery, nil
}

// Formats the WMTS query keys to lowercase, no WMTS query keys will be ignored
func splitQueryKeys(query url.Values, filter []string) (url.Values, url.Values) {
	newWMTSQuery := url.Values{}
	noneWMTSQuery := url.Values{}

	for key, values := range query {
		for _, wmtskey := range filter {
			if wmtskey == strings.ToLower(key) {
				newWMTSQuery[wmtskey] = values
			}
		}
	}

	for key, values := range query {
		isfilter := false
		for _, wmtskey := range filter {
			if wmtskey == strings.ToLower(key) {
				isfilter = true
				break
			}
		}
		if !isfilter {
			noneWMTSQuery[key] = values
		}
	}
	return newWMTSQuery, noneWMTSQuery
}

// formatKeysToQueryString takes a map and builds a query string
func formatKeysToQueryString(query url.Values) string {
	querystring := ""
	for key, values := range query {

		for _, v := range values {
			querystring = querystring + fmt.Sprintf("%s=%s&", key, v)
		}
	}
	return strings.TrimRight(querystring, "&")
}

// missingKeys checks if there are key-value pairs missing
// based on the array of keys provided
func missingKeys(query url.Values, keys []string) Exception {
	for _, v := range keys {
		if query[v] == nil {
			return MissingParameterValue(v)
		}
	}
	return nil
}

// ProcessRequest checks the quality of the request
// and if it's valid to process as a WMTS request
func ProcessRequest(config *Config, w http.ResponseWriter, r *http.Request) bool {

	// check if it's a WMTS request
	query, err := keysToLowerAndFilter(r.URL.Query())
	if err != nil {
		SendError(err, w, r)
		return false
	} else if len(query["service"]) < 1 || len(query["request"]) < 1 {
		return true
	} else if len(query["service"]) > 0 && strings.ToLower(query["service"][0]) != "wmts" {
		SendError(UnknownService(query["service"][0]), w, r)
		return false
	}

	// check what WMTS request and process
	switch strings.ToLower(query["request"][0]) {
	case "gettile":
		err := ProcessGetTileRequest(w, r)
		if err != nil {
			SendError(err, w, r)
			return false
		}
		return true
	case "getcapabilities":
		if len(config.Host) < 1 {
			return true
		}
		err := ProcessGetCapabilitiesRequest(config, w, r)
		if err != nil {
			SendError(err, w, r)
			return false
		}
		return false
	default:
		return true
	}
}
