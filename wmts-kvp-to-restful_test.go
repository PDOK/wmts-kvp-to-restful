package main

import (
	"reflect"
	"testing"
)

func TestFormatQueryKeysOther(t *testing.T) {
	a := "a"
	expected := "?a=a"

	query := map[string][]string{"a": {a}}

	formatKeys, other, _ := formatQueryKeys(query)

	for k, v := range query {
		if !reflect.DeepEqual(v, query[k]) {
			t.Errorf("Request was incorrect, got: %s, want: %s", formatKeys, query)
		}
	}

	if other != expected {
		t.Errorf("Expected none WMTS keys, got: %s, want: %s", other, expected)
	}
}

func TestFormatQueryKeys(t *testing.T) {
	a := "a"

	query := map[string][]string{"a": {a}}

	formatKeys, _, _ := formatQueryKeys(query)

	for k, v := range query {
		if !reflect.DeepEqual(v, query[k]) {

			t.Errorf("Request was incorrect, got: %s, want: %s", formatKeys, query)
		}
	}
}

func TestFormatWMTSQueryKeys(t *testing.T) {
	layer := "testlayer"
	tilematrixset := "EPSG:28992"
	tilematrix := "4"
	tilecol := "5"
	tilerow := "5"
	format := "image/png"

	query := map[string][]string{"layer": {layer}, "tilematrixset": {tilematrixset}, "tilematrix": {tilematrix}, "tilecol": {tilecol}, "tilerow": {tilerow}, "format": {format}}

	formatKeys, _, _ := formatQueryKeys(query)

	for k, v := range query {
		if !reflect.DeepEqual(v, query[k]) {

			t.Errorf("Request was incorrect, got: %s, want: %s", formatKeys, query)
		}
	}
}
