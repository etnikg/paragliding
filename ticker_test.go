package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

///Ticker tests

/////////////////Other testing functions

func Test_latestTimestamp(t *testing.T) {
	igcTracks := []tracks{
		tracks{TimeRecorded: time.Date(2018, 4, 25, 12, 32, 1, 0, time.UTC)},
		tracks{TimeRecorded: time.Now()},
		tracks{TimeRecorded: time.Date(2019, 4, 25, 12, 32, 1, 0, time.UTC)},
	}

	latestTS := latestTimestamp(igcTracks)
	if latestTS != igcTracks[2].TimeRecorded {
		t.Error("Not the latest timestamp")
	}
}

func Test_oldestTimestamp(t *testing.T) {
	igcTracks := []tracks{
		tracks{TimeRecorded: time.Date(2018, 4, 25, 12, 32, 1, 0, time.UTC)},
		tracks{TimeRecorded: time.Now()},
		tracks{TimeRecorded: time.Date(2019, 4, 25, 12, 32, 1, 0, time.UTC)},
	}

	oldestTS := oldestTimestamp(igcTracks)
	if oldestTS != igcTracks[0].TimeRecorded {
		t.Error("Not the oldest timestamp")
	}
}

func Test_oldestNewerTimestamp(t *testing.T) {
	igcTracks := []tracks{
		tracks{TimeRecorded: time.Date(2018, 4, 25, 12, 32, 1, 0, time.UTC)},
		tracks{TimeRecorded: time.Date(2018, 4, 26, 12, 32, 1, 0, time.UTC)},
		tracks{TimeRecorded: time.Date(2019, 4, 25, 12, 32, 1, 0, time.UTC)},
	}

	oldestNewTS := oldestNewerTimestamp("25.04.2018 12:34:30.314", igcTracks)

	if oldestNewTS != igcTracks[1].TimeRecorded {
		t.Error("Not the right timestamp")
	}
}

func Test_tickerTimestamps(t *testing.T) {
	igcTracks := []tracks{
		tracks{TimeRecorded: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	// No connection to the DB :(

	tickerTS := tickerTimestamps("25.04.2018 12:34:30.314")

	if tickerTS.oldestTimestamp != igcTracks[0].TimeRecorded {
		t.Error("Not the right timestamp")
	}
	if tickerTS.oldestNewerTimestamp != igcTracks[0].TimeRecorded {
		t.Error("Not the right timestamp")
	}
	if tickerTS.latestTimestamp != igcTracks[0].TimeRecorded {
		t.Error("Not the right timestamp")
	}
}

func Test_getAPITickerLatest(t *testing.T) {
	// instantiate mock HTTP server (just for the purpose of testing
	ts := httptest.NewServer(http.HandlerFunc(handlerTickerLatest))
	defer ts.Close()

	//create a request to our mock HTTP server
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the GET request, %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error executing the GET request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusOK %d, received %d. ", http.StatusOK, resp.StatusCode)
		return
	}

	req, err = http.NewRequest(http.MethodPost, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the POST request, %s", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound %d, received %d. ", http.StatusNotFound, resp.StatusCode)
		return
	}

}

func Test_getAPITicker(t *testing.T) {
	// instantiate mock HTTP server (just for the purpose of testing
	ts := httptest.NewServer(http.HandlerFunc(handlerTicker))
	defer ts.Close()

	//create a request to our mock HTTP server
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the GET request, %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error executing the GET request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusOK %d, received %d. ", http.StatusOK, resp.StatusCode)
		return
	}

	req, err = http.NewRequest(http.MethodPost, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the POST request, %s", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected StatusNotFound %d, received %d. ", http.StatusNotFound, resp.StatusCode)
		return
	}

}

func Test_getAPITickerTimestamp(t *testing.T) {
	// instantiate mock HTTP server (just for the purpose of testing
	ts := httptest.NewServer(http.HandlerFunc(handlerTickerTimestamp))
	defer ts.Close()

	//create a request to our mock HTTP server
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the GET request, %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error executing the GET request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected StatusBadRequest %d, received %d. ", http.StatusBadRequest, resp.StatusCode)
		return
	}

}
