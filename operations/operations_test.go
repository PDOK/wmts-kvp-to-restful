package operations

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func getKeys(i url.Values) []string {
	var results []string
	for k := range i {
		results = append(results, k)
	}
	return results
}

func getBodyAsString(b io.ReadCloser) string {
	bodyBytes, err := ioutil.ReadAll(b)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyBytes)
}

func TestKeysToLowerAndFilter(t *testing.T) {
	input := map[string][]string{"A": {"B"}, "D": {"e"}}
	expected := map[string][]string{"a": {"B"}, "d": {"e"}}
	result, err := keysToLowerAndFilter(input)

	if err != nil {
		t.Errorf("Got an error: %s", err)
	}

	for k, v := range expected {
		if result[k][0] != v[0] {
			t.Errorf("Expected %s but was not, got: %s", result[k][0], v[0])
		}
	}
}

func TestKeysToLowerAndFilterDouble(t *testing.T) {
	input := map[string][]string{"A": {"B", "1"}}
	result, err := keysToLowerAndFilter(input)

	if result != nil {
		t.Errorf("Got an error: %s", err)
	}

	if !strings.Contains(err.Error(), "A: B,1") {
		t.Errorf("Got an error: %s", err)
	}
}

func TestSplitQueryKeys(t *testing.T) {
	input := map[string][]string{"a": {"B", "1"}, "c": {"D", "2"}, "e": {"F", "3"}}
	filter := []string{"a", "g"}
	partone, parttwo := splitQueryKeys(input, filter)

	if !reflect.DeepEqual(getKeys(partone), []string{"a"}) {
		t.Errorf("Got an error: %s", getKeys(partone))
	}

	for k, v := range map[string][]string{"a": {"B", "1"}} {
		if partone[k][0] != v[0] {
			t.Errorf("Expected %s but was not, got: %s", partone[k][0], v[0])
		}
	}

	for k, v := range map[string][]string{"c": {"D", "2"}, "e": {"F", "3"}} {
		if parttwo[k][0] != v[0] {
			t.Errorf("Expected %s but was not, got: %s", parttwo[k][0], v[0])
		}
	}
}

func TestFormatKeysToQueryString(t *testing.T) {
	input := map[string][]string{"a": {"B", "1"}, "c": {"D", "2"}, "E": {"3"}}
	result := formatKeysToQueryString(input)

	if !strings.Contains(result, "a=B") {
		t.Errorf("Excepted a=B, but got: %s", result)
	}
	if !strings.Contains(result, "a=1") {
		t.Errorf("Excepted a=1, but got: %s", result)
	}
	if !strings.Contains(result, "c=D") {
		t.Errorf("Excepted c=D, but got: %s", result)
	}
	if !strings.Contains(result, "c=2") {
		t.Errorf("Excepted c=2, but got: %s", result)
	}
	if !strings.Contains(result, "E=3") {
		t.Errorf("Excepted E=3, but got: %s", result)
	}
}

func TestMissingKeys(t *testing.T) {
	input := map[string][]string{"a": {"B", "1"}, "c": {"D", "2"}, "E": {"3"}}
	keys := []string{"a", "c"}
	result := missingKeys(input, keys)

	if result != nil {
		t.Errorf("Excepted keys %s, but got: %s", keys, result)
	}
}

func TestMissingKeysExcpetion(t *testing.T) {
	input := map[string][]string{"a": {"B", "1"}, "c": {"D", "2"}, "E": {"3"}}
	keys := []string{"a", "d"}
	result := missingKeys(input, keys)

	if !strings.Contains(result.Error(), "Missing parameter: d") {
		t.Errorf("Excepted error `Missing parameter: d`, but got: %s", result)
	}
}

func TestProcesRequestNoWMTS(t *testing.T) {
	var proxy bool
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Path: "local", RawQuery: "nowmts=nowmts"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	config := &Config{Host: "localhost", Template: "testTemplate"}
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy = ProcesRequest(config, w, mockRequest)
		}))
	defer ts.Close()

	http.Get(ts.URL)

	if !proxy {
		t.Errorf("Expected %t but was not, got: %t", true, proxy)
	}
}

func TestProcesRequest(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Path: "local", RawQuery: "service=wbs&service=wks"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	config := &Config{Host: "localhost", Template: "testTemplate"}
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = ProcesRequest(config, w, mockRequest)
		}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	expected := "Multiple query values found for service: wbs,wks"
	body := getBodyAsString(resp.Body)
	if !strings.Contains(body, expected) {
		t.Errorf("Expected %s but was not, got: %s", expected, body)
	}
}

func TestProcesRequestUnknownService(t *testing.T) {
	var mockRequest = &http.Request{
		Method:     "GET",
		Host:       "example.com",
		URL:        &url.URL{Path: "local", RawQuery: "service=wbs&request=getcapabilities"},
		Header:     http.Header{},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		RemoteAddr: "192.0.2.1:1234",
	}
	config := &Config{Host: "localhost", Template: "testTemplate"}
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = ProcesRequest(config, w, mockRequest)
		}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	expected := "Missing SERVICE key or incorrect value, found: wbs"
	body := getBodyAsString(resp.Body)
	if !strings.Contains(body, expected) {
		t.Errorf("Expected %s but was not, got: %s", expected, body)
	}
}
