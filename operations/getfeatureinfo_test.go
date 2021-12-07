package operations

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseInfoFormat(t *testing.T) {
	format := "plain/text"
	expected := ".txt"
	result, err := parseFileExtension(format)
	if err != nil || result != expected {
		t.Errorf("Expected format is: %s but was: %s", expected, result)
	}
}

func TestParseInvalidInfoFormat(t *testing.T) {
	format := "image/png"
	expected := ""
	result, err := parseFileExtension(format)
	if result != "" || err.Code() != "InvalidParameterValue" {
		t.Errorf("Expected format is: %s but was: %s", expected, err.Code())
	}
}

func TestProcessGetFeatureInfoRequest(t *testing.T) {
	var mockRequest = &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{Path: "local", RawQuery: "service=WMTS&request=GetFeatureInfo&version=1.0.0" +
			"&layer=achtergrondvisualisatie&tilematrixset=EPSG:28992&tilematrix=14" +
			"&tilecol=col&tilerow=row&infoformat=text/plain&j=1&i=2&testkey=testvalue"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	expected := "local/achtergrondvisualisatie/EPSG:28992/14/row/col/1/2.txt"
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProcessGetFeatureInfoRequest(w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if mockRequest.URL.String() != expected {
		t.Errorf("Expected %s but was not, got: %s", expected, mockRequest.URL.String())
	}
}

func TestProcessGetFeatureInfoRequestMissingKeys(t *testing.T) {
	var err Exception
	var mockRequest = &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{Path: "local", RawQuery: "service=WMTS&request=GetFeatureInfo&version=1.0.0" +
			"&layer=achtergrondvisualisatie&tilematrix=14" +
			"&tilecol=col&tilerow=row&infoformat=text/plain&j=1&i=2&testkey=testvalue"},
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
