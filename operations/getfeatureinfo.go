package operations

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
)

var getFeatureInfoRegex = regexp.MustCompile(`^.*:(.*)$`)

// Although the spc indices that TileRow must supersede TileCol, this does not seem to work.
const getFeatureInfoRestTemplate = `/{{ .Layer }}/{{ .TileMatrixSet }}/{{ .TileMatrix }}/{{ .TileCol }}/{{ .TileRow }}/{{ .I }}/{{ .J }}{{ .FileExtension }}`

// ProcessGetFeatureInfoRequest - Translates KVP requests to RestFUL requests
func ProcessGetFeatureInfoRequest(w http.ResponseWriter, r *http.Request) Exception {
	wmtskeys, otherkeys := splitQueryKeys(r.URL.Query(), getFeatureInfoKeys())
	err := missingKeys(wmtskeys, getFeatureInfoKeys())
	if err != nil {
		return err
	}

	url, err := getFeatureInfoQueryToPath(wmtskeys)
	if err != nil {
		return err
	}

	r.URL.Path = strings.TrimRight(r.URL.Path, "/") + url
	if len(otherkeys) > 0 {
		r.URL.RawQuery = formatKeysToQueryString(otherkeys)
	} else {
		r.URL.RawQuery = ""
	}
	return nil
}

func getFeatureInfoQueryToPath(query url.Values) (string, Exception) {

	type RestParameters struct {
		Layer         string
		TileMatrixSet string
		TileMatrix    string
		TileCol       string
		TileRow       string
		I             string
		J             string
		FileExtension string
	}

	tilematrix := query["tilematrix"][0]
	groups := getFeatureInfoRegex.FindAllStringSubmatch(tilematrix, -1)
	if groups != nil {
		tilematrix = groups[0][1]
	}

	fileExtension, err := parseFileExtension(query["infoformat"][0])
	if err != nil {
		return "", err
	}

	restParameters := &RestParameters{Layer: query["layer"][0], TileMatrixSet: query["tilematrixset"][0],
		TileMatrix: tilematrix, TileCol: query["tilecol"][0], TileRow: query["tilerow"][0],
		I: query["i"][0], J: query["j"][0], FileExtension: fileExtension}

	buf := new(bytes.Buffer)
	template.Must(template.New("getFeatureInfoResttemplate").Parse(getFeatureInfoRestTemplate)).Execute(buf, restParameters)

	return buf.String(), nil
}

func parseFileExtension(format string) (string, Exception) {

	fileExtension := format
	switch fileExtension {
	case "plain/text":
		fileExtension = ".txt"
	case "text/html":
		fileExtension = ".html"
	case "application/json":
		fileExtension = ".json"
	case "text/xml":
		fileExtension = ".xml"
	default:
		return "", InvalidParameterValue("infoformat", fileExtension)
	}
	return fileExtension, nil
}

// GetCapabilitiesKeys list of manitory WMTS gettile key value pairs
func getFeatureInfoKeys() []string {
	return []string{"service", "request", "version", "layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "i", "j", "infoformat"}
}
