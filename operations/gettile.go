package operations

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"text/template"
)

var regex = regexp.MustCompile(`^.*:(.*)$`)

const restTemplate = `/{{ .Layer }}/{{ .Tilematrixset }}/{{ .Tilematrix }}/{{ .Tilecol }}/{{ .Tilerow }}{{ .Fileextension }}`

func tileQueryToPath(query url.Values) string {

	type RestParameters struct {
		Layer         string
		Tilematrixset string
		Tilematrix    string
		Tilecol       string
		Tilerow       string
		Fileextension string
	}

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

// GetCapabilitiesKeys list of manitory WMTS gettile key value pairs
func getTileKeys() []string {
	return []string{"layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}
}

// ProcessGetTileRequest rewrites the KVP request as RestFUL
// and alters the request so it can be proxied
func ProcessGetTileRequest(w http.ResponseWriter, r *http.Request) Exception {
	wmtskeys, otherkeys := splitQueryKeys(r.URL.Query(), getTileKeys())
	err := missingKeys(wmtskeys, getTileKeys())
	if err != nil {
		return err
	}

	r.URL.Path = strings.TrimRight(r.URL.Path, "/") + tileQueryToPath(wmtskeys)
	if len(otherkeys) > 0 {
		r.URL.RawQuery = formatKeysToQueryString(otherkeys)
	} else {
		r.URL.RawQuery = ""
	}
	return nil
}
