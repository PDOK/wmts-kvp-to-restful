package operations

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestProcessGetTileRequest(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Path: "local", RawQuery: "service=WMTS&request=GetTile&version=1.0.0&layer=a&tilematrixset=b&tilematrix=c&tilecol=d&tilerow=e&format=f&testkey=testvalue"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	expected := "local/a/b/c/d/e.png?testkey=testvalue"
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProcessGetTileRequest(w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if mockRequest.URL.String() != expected {
		t.Errorf("Expected %s but was not, got: %s", expected, mockRequest.URL.String())
	}
}

func TestProcessGetTileRequestLongerPath(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "subdomain.example.org",
		URL:        &url.URL{Path: "local/a/path", RawQuery: "service=WMTS&request=GetTile&version=1.0.0&layer=e&tilematrixset=d&tilematrix=d:c&tilecol=b&tilerow=a&format=image/png&testkey=testvalue"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	expected := "local/a/path/e/d/c/b/a.png?testkey=testvalue"
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProcessGetTileRequest(w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if mockRequest.URL.String() != expected {
		t.Errorf("Expected %s but was not, got: %s", expected, mockRequest.URL.String())
	}
}

func TestProcessGetTileRequestMissingKeys(t *testing.T) {
	var err Exception
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Path: "local", RawQuery: "service=WMTS&request=GetTile&version=1.0.0&layer=a&tilematrix=c&tilecol=d&tilerow=e&format=f"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	expected := "Missing parameter: tilematrixset"
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = ProcessGetTileRequest(w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected %s but was not, got: %s", expected, err.Error())
	}
}

func TestProcessGetTileRequestNoRawQuery(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Path: "local", RawQuery: "service=WMTS&request=GetTile&version=1.0.0&layer=a&tilematrixset=b&tilematrix=c&tilecol=d&tilerow=e&format=f"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	expected := "local/a/b/c/d/e.png"
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProcessGetTileRequest(w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if mockRequest.URL.String() != expected {
		t.Errorf("Expected %s but was not, got: %s", expected, mockRequest.URL.String())
	}
}

func TestGetTileKeys(t *testing.T) {
	expected := []string{"service", "request", "version", "layer", "tilematrixset", "tilematrix", "tilecol", "tilerow", "format"}
	if !reflect.DeepEqual(getTileKeys(), expected) {
		t.Errorf("Expected %s but was not, got: %s", expected, getTileKeys())
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
