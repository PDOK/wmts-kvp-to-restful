package main

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestNormalPath(t *testing.T) {
	partone := "part/one"
	parttwo := "/part/two"
	newpath := buildNewPath(partone, parttwo)
	expectednewpath := partone + parttwo

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	}
}

func TestPathWithTrailingSlashes(t *testing.T) {
	partone := "part/one////"
	parttwo := "/part/two"
	newpath := buildNewPath(partone, parttwo)
	expectednewpath := "part/one/part/two"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
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

func TestCompleteTileQuery(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "tilerow": {"e"}, "format": {"f"},
	}

	expectedResult := [0]string{}
	var result [0]string
	copy(result[:], findMissingParams(query, GetTileKeys[:]))

	if result != expectedResult {
		t.Errorf("Complete query was found incomplete.")
	}
}

func TestTileQueryCaseInsensitive(t *testing.T) {
	query := map[string][]string{
		"LAYER": {"a"}, "TILEMATRIXSET": {"b"}, "tileMATRIX": {"c"}, "TILECOL": {"d"}, "TILErow": {"e"}, "format": {"f"},
	}

	expectedResult := [0]string{}
	var result [0]string
	copy(result[:], findMissingParams(query, GetTileKeys[:]))

	if result != expectedResult {
		t.Errorf("Tile query check does not handle uppercase correctly.")
	}
}

func TestNotATileQuery(t *testing.T) {
	query := map[string][]string{"request": {"getCapabilities"}}

	expectedResult := [0]string{}
	var result [0]string
	copy(result[:], findMissingParams(query, GetTileKeys[:]))

	if result != expectedResult {
		t.Errorf("Query was not a tile query but was identified as such.")
	}
}

func TestIncompleteTileQuery(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	expectedResult := [1]string{"tilerow"}

	var result [1]string
	copy(result[:], findMissingParams(query, GetTileKeys[:]))

	if result != expectedResult {
		t.Errorf("Incomplete query was found complete.")
	}
}

func TestGetTileOperation(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"}, "request": {"GetTile"},
	}

	operation, exception := getOperation(query)

	if operation != GetTile {
		t.Errorf("Instead of GetTile, the found operation was: " + string(operation))
	} else if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception.Error(), "nil")
	}
}

func TestMultipleRequestOperation(t *testing.T) {
	query := map[string][]string{
		"request": {"GetTile", "GetTile", "GetTile", "GetCapabilities", "GetTile", "GetTile", "GetTile", "GetFeatureInfo", "GetFeatureInfo"},
	}
	operation, _ := getOperation(query)
	if operation != None {
		t.Errorf("Instead of None, the found operation was: " + string(operation))
	}
}

func TestMultipleRequestOperationException(t *testing.T) {
	query := map[string][]string{
		"request": {"GetTile", "GetTile", "GetTile", "GetCapabilities", "GetTile", "GetTile", "GetTile", "GetFeatureInfo", "GetFeatureInfo"},
	}
	_, exception := getOperation(query)
	if exception == nil {
		t.Errorf(`Exception was incorrect, got: nil, want: "multiple query values found".`)
	} else if exception.Error() != ExInvalidRequestValues {
		t.Errorf(`Exception was incorrect, got: %s, want: %s.`, exception.Error(), ExInvalidRequestValues)
	}
}

func TestGetFeatureInfoOperation(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"}, "request": {"GetFeatureInfo"},
	}

	operation, _ := getOperation(query)

	if operation != GetFeatureInfo {
		t.Errorf("Instead of GetFeatureInfo, the found operation was: " + string(operation))
	}
}

func TestGetFeatureInfoOperationException(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"}, "request": {"GetFeatureInfo"},
	}

	_, exception := getOperation(query)

	if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception.Error(), "nil")
	}
}

func TestMissingOperation(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	operation, _ := getOperation(query)

	if operation != None {
		t.Errorf("Instead of None, the found operation was: " + string(operation))
	}
}

func TestMissingOperationException(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	_, exception := getOperation(query)

	if exception == nil {
		t.Errorf(`Exception was incorrect, got: nil, want: "multiple query values found".`)
	} else if exception.Error() != ExInvalidRequestValues {
		t.Errorf(`Exception was incorrect, got: %s, want: %s.`, exception.Error(), ExInvalidRequestValues)
	}
}
func TestFormatQuery(t *testing.T) {
	query := url.Values{"REQUEST": {"GetFeatureInfo"}, "LAYER": {"a"}}
	expectedQuery := url.Values{"request": {"GetFeatureInfo"}, "layer": {"a"}}

	resultQuery, _ := formatQuery(query)
	if !reflect.DeepEqual(expectedQuery, resultQuery) {
		t.Errorf("Query keys were not lowercased.")
	}
}

func TestFormatQueryException(t *testing.T) {
	query := url.Values{"REQUEST": {"GetFeatureInfo"}, "LAYER": {"a"}}
	_, exception := formatQuery(query)
	if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception.Error(), "nil")
	}
}

func TestFormatQueryRaisesErrorOnMultipleRequestValues(t *testing.T) {
	query := url.Values{"REQUEST": {"GetFeatureInfo", "y", "x"}, "LAYER": {"a"}}

	resultQuery, _ := formatQuery(query)
	if resultQuery != nil {
		t.Errorf("With erroneous input result query should be nil.")
	}
}

func TestFormatQueryRaisesErrorOnMultipleRequestValuesException(t *testing.T) {
	query := url.Values{"REQUEST": {"GetFeatureInfo", "y", "x"}, "LAYER": {"a"}}

	_, exception := formatQuery(query)
	if exception == nil {
		t.Errorf(`Exception was incorrect, got: nil, want: %s.`, ExMultipleValuesFound)
	} else if exception.Error() != ExMultipleValuesFound {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception.Error(), "nil")
	}
}

func TestOperationFromString(t *testing.T) {
	result := operationFromString("getfeatureinfo")
	if result != GetFeatureInfo {
		t.Errorf(`String "getfeatureinfo" was not converted to Operation GetFeatureInfo. Instead we got: ` + string(result))
	}
}

// --------- HandleOperation tests --------------------
var mockRequest = &http.Request{
	Method:     "GET",
	Host:       "example.com",
	URL:        &url.URL{Path: "http://test/"},
	Header:     http.Header{},
	Proto:      "HTTP/1.1",
	ProtoMajor: 1,
	ProtoMinor: 1,
	RemoteAddr: "192.0.2.1:1234",
	RequestURI: "/test/",
}

func TestHandleOperationIncomingError(t *testing.T) {
	// handleOperation
	incomingException := errors.New("test")
	_, _, _, _, exception := handleOperation(url.Values{}, mockRequest, incomingException)

	if exception != incomingException {
		t.Errorf(`Exception was incorrect, got: %s, want: %s.`, exception.Error(), "test")
	}
}

func TestHandleOperationIncomingErrorContentType(t *testing.T) {
	// handleOperation
	incomingException := errors.New("test")
	var expectedContentType = "application/xml; charset=UTF-8"
	_, _, contentType, _, _ := handleOperation(url.Values{}, mockRequest, incomingException)
	if contentType != expectedContentType {
		t.Errorf("Contenttype was incorrect, got: %s, want: %s.", contentType, expectedContentType)
	}
}

func TestHandleOperationIncomingErrorStatusCode(t *testing.T) {
	// handleOperation
	incomingException := errors.New("test")
	statusCode, _, _, _, _ := handleOperation(url.Values{}, mockRequest, incomingException)
	if statusCode != http.StatusBadRequest {
		t.Errorf("With error response statusCode should be %d found %d", http.StatusBadRequest, statusCode)
	}
}

func TestHandleOperationNoOperationError(t *testing.T) {
	// handleOperation
	_, _, _, _, exception := handleOperation(url.Values{}, mockRequest, nil)

	if exception.Error() != ExInvalidRequestValues {
		t.Errorf(`Exception was incorrect, got: %s, want: %s.`, exception.Error(), ExInvalidRequestValues)
	}
}

func TestHandleOperationNoOperationErrorContentType(t *testing.T) {
	// handleOperation
	var expectedContentType = "application/xml; charset=UTF-8"
	_, _, contentType, _, _ := handleOperation(url.Values{}, mockRequest, nil)
	if contentType != expectedContentType {
		t.Errorf("Contenttype was incorrect, got: %s, want: %s.", contentType, expectedContentType)
	}
}

func TestHandleOperationNoOperationErrorStatusCode(t *testing.T) {
	// handleOperation
	statusCode, _, _, _, _ := handleOperation(url.Values{}, mockRequest, nil)
	if statusCode != http.StatusBadRequest {
		t.Errorf("With error response statusCode should be %d found %d", http.StatusBadRequest, statusCode)
	}
}

func TestHandleOperationGetTileStatusCode(t *testing.T) {
	// handleOperation
	var query = url.Values{
		"layer": {"brta"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"1"}, "tilerow": {"1"},
		"format": {"image/jpeg"}, "request": {"GetTile"}}
	statusCode, _, _, _, _ := handleOperation(query, mockRequest, nil)
	if statusCode != http.StatusOK {
		t.Errorf("With default response statusCode should be %d found %d", http.StatusOK, statusCode)
	}
}

func TestHandleOperationGetTileContentType(t *testing.T) {
	// handleOperation
	var query = url.Values{
		"layer": {"brta"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"1"}, "tilerow": {"1"},
		"format": {"image/jpeg"}, "request": {"GetTile"}}
	var expectedContentType = ""
	_, _, contentType, _, _ := handleOperation(query, mockRequest, nil)
	if contentType != expectedContentType {
		t.Errorf("Contenttype was incorrect, got: %s, want: %s.", contentType, expectedContentType)
	}
}

func TestHandleOperationGetCapabilitiesUrl(t *testing.T) {
	// handleOperation
	var query = url.Values{
		"layer": {"brta"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"1"}, "tilerow": {"1"},
		"format": {"image/jpeg"}, "request": {"GetCapabilities"}}
	expectedURL := "http://test/v1_0/WMTSCapabilities.xml"
	_, path, _, _, _ := handleOperation(query, mockRequest, nil)
	if path != expectedURL {
		t.Errorf("GetCapabilities url was incorrect, got: %s, want: %s.", path, expectedURL)
	}
}

func TestHandleOperationGetCapabilitiesException(t *testing.T) {
	// handleOperation
	var query = url.Values{
		"layer": {"brta"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"1"}, "tilerow": {"1"},
		"format": {"image/jpeg"}, "request": {"GetCapabilities"}}
	_, _, _, _, exception := handleOperation(query, mockRequest, nil)

	if exception != nil {
		t.Errorf(`Exception was incorrect, got: %s, want: %s.`, exception.Error(), "nil")
	}
}

func TestHandleOperationGetCapabilitiesStatusCode(t *testing.T) {
	// handleOperation
	var query = url.Values{
		"layer": {"brta"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"1"}, "tilerow": {"1"},
		"format": {"image/jpeg"}, "request": {"GetCapabilities"}}
	statusCode, _, _, _, _ := handleOperation(query, mockRequest, nil)
	if statusCode != http.StatusOK {
		t.Errorf("With default response statusCode should be %d found %d", http.StatusOK, statusCode)
	}
}
