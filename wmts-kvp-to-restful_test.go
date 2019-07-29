package main

import (
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

func TestPathWithTrailingSlashs(t *testing.T) {
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

	newpath, exception := queryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	} else if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception, "nil")
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

	newpath, exception := queryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	} else if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception, "nil")
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

	newpath, exception := queryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	} else if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception, "nil")
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

	newpath, exception := queryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".png"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	} else if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception, "nil")
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

	newpath, exception := queryToPath(query)
	expectednewpath := "/" + layer + "/" + tilematrixset + "/" + tilematrix + "/" + tilecol + "/" + tilerow + ".jpeg"

	if newpath != expectednewpath {
		t.Errorf("Request was incorrect, got: %s, want: %s.", newpath, expectednewpath)
	} else if exception != nil {
		t.Errorf("Exception was incorrect, got: %s, want: %s.", exception, "nil")
	}
}

func TestCompleteTileQuery(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "tilerow": {"e"}, "format": {"f"},
	}

	expectedResult := [0]string{}
	var result [0]string
	copy(result[:], validateTileQuery(query))

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
	copy(result[:], validateTileQuery(query))

	if result != expectedResult {
		t.Errorf("Tile query check does not handle uppercase correctly.")
	}
}

func TestNotATileQuery(t *testing.T) {
	query := map[string][]string{"request": {"getCapabilities"}}

	expectedResult := [0]string{}
	var result [0]string
	copy(result[:], validateTileQuery(query))

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
	copy(result[:], validateTileQuery(query))

	if result != expectedResult {
		t.Errorf("Incomplete query was found complete.")
	}
}

func TestGetTileOperation(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"}, "request": {"GetTile"},
	}

	operation := getOperation(query)

	if operation != GetTile {
		t.Errorf("Instead of GetTile, the found operation was: " + string(operation))
	}
}

func TestMultipleRequestOperation(t *testing.T) {
	query := map[string][]string{
		"request": {"GetTile", "GetTile", "GetTile", "GetCapabilities", "GetTile", "GetTile", "GetTile", "GetFeatureInfo", "GetFeatureInfo"},
	}

	operation := getOperation(query)

	if operation != GetCapabilities {
		t.Errorf("Instead of GetCapabilities, the found operation was: " + string(operation))
	}
}

func TestGetFeatureInfoOperation(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"}, "request": {"GetFeatureInfo"},
	}

	operation := getOperation(query)

	if operation != GetFeatureInfo {
		t.Errorf("Instead of GetFeatureInfo, the found operation was: " + string(operation))
	}
}

func TestMissingOperation(t *testing.T) {
	query := map[string][]string{
		"layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	operation := getOperation(query)

	if operation != None {
		t.Errorf("Instead of None, the found operation was: " + string(operation))
	}
}

func TestMultipleOperationsGetCapabilities(t *testing.T) {
	query := map[string][]string{
		"request": {"GetFeatureInfo", "GetTile", "GetCapabilities"}, "layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	operation := getOperation(query)

	if operation != GetCapabilities {
		t.Errorf("Instead of GetCapabilities, the found operation was: " + string(operation))
	}
}

func TestMultipleOperationsGetTiles(t *testing.T) {
	query := map[string][]string{
		"request": {"GetFeatureInfo", "GetTile", "x"}, "layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	operation := getOperation(query)

	if operation != GetTile {
		t.Errorf("Instead of GetTile, the found operation was: " + string(operation))
	}
}

func TestMultipleOperationsGetFeatureInfo(t *testing.T) {
	query := map[string][]string{
		"request": {"GetFeatureInfo", "y", "x"}, "layer": {"a"}, "tilematrixset": {"b"}, "tilematrix": {"c"}, "tilecol": {"d"}, "format": {"f"},
	}

	operation := getOperation(query)

	if operation != GetFeatureInfo {
		t.Errorf("Instead of GetFeatureInfo, the found operation was: " + string(operation))
	}
}
