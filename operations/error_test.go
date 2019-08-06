package operations

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendError(t *testing.T) {

	code := "MissingParameterValue"
	status := 400
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := WMTSException{ErrorMessage: fmt.Sprintf("Missing parameters: " + strings.Join([]string{"Request", "Service"}, ", ")), ErrorCode: code, StatusCode: status}
			SendError(err, w, r)
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

	if !strings.Contains(bodyString, code) {
		t.Errorf("Expected %s but was not, got: %s", code, bodyString)
	}
	if resp.StatusCode != status {
		t.Errorf("Expected statuscode %d but was not, got: %d", status, resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestFindMissingParamsHappyFlow(t *testing.T) {
	correctQuery := map[string][]string{"1": {"1"}, "2": {"1"}, "3": {"1"}, "4": {"1"}, "5": {"1"}, "6": {"1"}}
	expectedKeys := []string{"1", "2", "3", "4", "5", "6"}
	missingParameters := FindMissingParams(correctQuery, expectedKeys)

	if len(missingParameters) != 0 {
		t.Errorf("ExpectedKeys should be empty but was not, got: %s", strings.Join(missingParameters, ","))
	}
}

func TestFindMissingParamsMissing(t *testing.T) {

	correctQuery := map[string][]string{"1": {"1"}, "2": {"1"}, "3": {"1"}, "4": {"1"}, "5": {"1"}}
	expectedKeys := []string{"1", "2", "3", "4", "5", "6"}
	missingParameters := FindMissingParams(correctQuery, expectedKeys)

	if len(missingParameters) != 1 {
		t.Errorf("ExpectedKeys should be empty but was not, got: %s", strings.Join(missingParameters, ","))
	}
}

func TestUnknownService(t *testing.T) {

	s := "teststring"
	err := UnknownService(s)

	if !strings.Contains(err.Error(), s) {
		t.Errorf("Error should contain: %s, got: %s", s, err.Error())
	}
}

func TestMissingParameterValue(t *testing.T) {

	s := "teststring"
	err := MissingParameterValue(s)

	if !strings.Contains(err.Error(), s) {
		t.Errorf("Error should contain: %s, got: %s", s, err.Error())
	}
}
