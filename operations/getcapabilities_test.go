package operations

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestHostAndPath(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Host: "example.com", Path: "/example/path"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}

	result := hostAndPath(mockRequest)

	if result.Protocol != "http" {
		t.Errorf("Expected protocol: http, got: %s", result.Protocol)
	}
	if result.Host != "example.com" {
		t.Errorf("Expected host: example.com, got: %s", result.Host)
	}
	if result.Path != "/example/path" {
		t.Errorf("Expected path: /example/path, got: %s", result.Path)
	}
}

func TestHostAndPathHeaders(t *testing.T) {
	headers := http.Header{"X-Forwarded-Proto": {"https"}, "X-Forward-Host": {"new.example.org"}, "X-Script-Name": {"/new/example/path"}}
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Host: "example.com", Path: "/example/path"},
		Header:     headers,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	result := hostAndPath(mockRequest)

	if result.Protocol != "https" {
		t.Errorf("Expected protocol: https, got: %s", result.Protocol)
	}
	if result.Host != "new.example.org" {
		t.Errorf("Expected host: new.example.org, got: %s", result.Host)
	}
	if result.Path != "/new/example/path" {
		t.Errorf("Expected path: /new/example/path, got: %s", result.Path)
	}
}

func TestGetCapabilitiesTemplate(t *testing.T) {
	template, err := getCapabilitiesTemplate("testTemplate")

	if err != nil {
		t.Errorf("Expected template but got no template!")
	}

	if err == nil && fmt.Sprint(reflect.TypeOf(template)) != "*template.Template" {
		t.Errorf("Expected template but got: %s", reflect.TypeOf(template))
	}
}

func TestGetCapabilitiesKeys(t *testing.T) {
	expected := []string{"request", "service", "version"}
	result := getCapabilitiesKeys()

	for k, v := range expected {
		present := false
		for _, j := range result {

			if j == v {
				present = true
			}
		}
		if !present {
			t.Errorf("Expected %s but was not, got: %s", result[k], v)
		}
	}
}

func TestProcessGetCapabilitiesRequest(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Host: "example.com", Path: "/example/path"},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	config := &Config{Host: "localhost", Template: "testTemplate"}

	content := "http://example.com/example/path"
	status := 200
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProcessGetCapabilitiesRequest(config, w, mockRequest)
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
