package operations

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var mockRequest = &http.Request{
	Method:     "GET",
	Host:       "example.com",
	URL:        &url.URL{Path: "local"},
	Header:     http.Header{},
	Proto:      "HTTP/1.1",
	ProtoMajor: 1,
	ProtoMinor: 1,
	RemoteAddr: "192.0.2.1:1234",
}

func TestProcesGetTileRequest(t *testing.T) {
	expected := "local/a/b/c/d/e.png?testkey=testvalue"
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := map[string][]string{"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "tilerow": {"e"}, "format": {"f"}}
			ProcesGetTileRequest(query, "testkey=testvalue", w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if mockRequest.URL.String() != expected {
		t.Errorf("Expected %s but was not, got: %s", expected, mockRequest.URL.String())
	}
}

func TestProcesGetCapabilitiesRequest(t *testing.T) {
	content := "http://localhost/"
	status := 200
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := map[string][]string{"request": {"a"}, "service": {"b"}}
			template := "testTemplate"
			ProcesGetCapabilitiesRequest(query, template, w, r)
		}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	if !strings.Contains(bodyString, content) {
		t.Errorf("Expected %s but was not, got: %s", content, bodyString)
	}
	if resp.StatusCode != status {
		t.Errorf("Expected statuscode %d but was not, got: %d", status, resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestProcesGetCapabilitiesRequestXForwardHeaders(t *testing.T) {
	content := "https://host.new/a/new/path"
	status := 200
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set("X-Forward-Host", "host.new")
			r.Header.Set("X-Script-Name", "/a/new/path")
			query := map[string][]string{"request": {"a"}, "service": {"b"}}
			template := "testTemplate"
			ProcesGetCapabilitiesRequest(query, template, w, r)
		}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	if !strings.Contains(bodyString, content) {
		t.Errorf("Expected %s but was not, got: %s", content, bodyString)
	}
	if resp.StatusCode != status {
		t.Errorf("Expected statuscode %d but was not, got: %d", status, resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestGetCapabilitiesKeys(t *testing.T) {
	expected := []string{"request", "service"}
	if !reflect.DeepEqual(GetCapabilitiesKeys(), expected) {
		t.Errorf("Expected %s but was not, got: %s", expected, GetCapabilitiesKeys())
	}
}

func TestGetTileKeys(t *testing.T) {
	expected := []string{"request", "service", "layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}
	if !reflect.DeepEqual(GetTileKeys(), expected) {
		t.Errorf("Expected %s but was not, got: %s", expected, GetTileKeys())
	}
}

func TestHappyFlow(t *testing.T) {
	layer := "testlayer"
	tilematrixset := "EPSG:28992"
	tilematrix := "4"
	tilecol := "5"
	tilerow := "5"
	format := "image/png"

	query := map[string][]string{"layer": {layer}, "tilematrixset": {tilematrixset}, "tilematrix": {tilematrix}, "tilecol": {tilecol}, "tilerow": {tilerow}, "format": {format}}

	newpath := tileQueryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	}
}

func TestGWCTileMatrixIssue(t *testing.T) {
	layer := "testlayer"
	tilematrixset := "EPSG:28992"
	gwctilematrixprefix := "EPSG:28992:"
	tilematrix := "4"
	tilecol := "5"
	tilerow := "5"
	format := "image/png"

	query := map[string][]string{"layer": {layer}, "tilematrixset": {tilematrixset}, "tilematrix": {gwctilematrixprefix + tilematrix}, "tilecol": {tilecol}, "tilerow": {tilerow}, "format": {format}}

	newpath := tileQueryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	}
}

func TestObscureGWCTileMatrixIssue(t *testing.T) {
	layer := "testlayer"
	tilematrixset := "EPSG:25831:RWS"
	gwctilematrixprefix := "EPSG:25831:RWS:"
	tilematrix := "4"
	tilecol := "5"
	tilerow := "5"
	format := "image/png"

	query := map[string][]string{"layer": {layer}, "tilematrixset": {tilematrixset}, "tilematrix": {gwctilematrixprefix + tilematrix}, "tilecol": {tilecol}, "tilerow": {tilerow}, "format": {format}}

	newpath := tileQueryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	}
}

func TestImagePng8FormatMapping(t *testing.T) {
	layer := "testlayer"
	tilematrixset := "EPSG:3857"
	tilematrix := "4"
	tilecol := "5"
	tilerow := "5"
	format := "image/png8"

	query := map[string][]string{"layer": {layer}, "tilematrixset": {tilematrixset}, "tilematrix": {tilematrix}, "tilecol": {tilecol}, "tilerow": {tilerow}, "format": {format}}

	newpath := tileQueryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	}
}

func TestImageJpegFormatMapping(t *testing.T) {
	layer := "testlayer"
	tilematrixset := "EPSG:3857"
	tilematrix := "4"
	tilecol := "5"
	tilerow := "5"
	format := "image/jpeg"

	query := map[string][]string{"layer": {layer}, "tilematrixset": {tilematrixset}, "tilematrix": {tilematrix}, "tilecol": {tilecol}, "tilerow": {tilerow}, "format": {format}}

	newpath := tileQueryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".jpeg"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	}
}
