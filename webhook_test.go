package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_webhookNewTrack_NotImplemented(t *testing.T) {

	// instantiate mock HTTP server (just for the purpose of testing
	ts := httptest.NewServer(http.HandlerFunc(webhookNewTrack))
	defer ts.Close()

	//create a request to our mock HTTP server
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the POST request, %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != 200 {
		t.Errorf("Expected StatusNotImplemented %d, received %d. ", 200, resp.StatusCode)
		return
	}

}
func Test_getNewWebhook(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(webhookNewTrack))
	defer ts.Close()

	client := &http.Client{}

	urlTest := Webhook{}
	urlTest.WebhookURL = "https://discordapp.com:8080/api/webhooks/504970988400803842/cyPUQQw0laWWVSikkV-cwvZKv97xyUkbi-2aDX2fJZccJYmORHOknS155L2lUX3_LPlM"

	jsonData, _ := json.Marshal(urlTest)

	request, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Errorf("Error making the POST request, %s", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusOK %d, received %d. ", http.StatusOK, resp.StatusCode)
		return
	}

	if resp.StatusCode == 400 {
		assert.Equal(t, 400, resp.StatusCode, "OK response is expected")
	} else {
		assert.Equal(t, 200, resp.StatusCode, "OK response is expected")
	}

}

func Test_getNewWebhookID(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(webhookNewTrack))
	defer ts.Close()

	client := &http.Client{}

	urlTest := Webhook{}
	urlTest.WebhookID = "0"

	jsonData, _ := json.Marshal(urlTest)

	request, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Errorf("Error making the POST request, %s", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	if resp.StatusCode == 400 {
		assert.Equal(t, 400, resp.StatusCode, "OK response is expected")
	} else {
		assert.Equal(t, 200, resp.StatusCode, "OK response is expected")
	}

}

func Test_webhookNewTrack(t *testing.T) {
	// instantiate mock HTTP server (just for the purpose of testing
	ts := httptest.NewServer(http.HandlerFunc(webhookNewTrack))
	defer ts.Close()

	//create a request to our mock HTTP server
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the POST request, %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusOK %d, received %d. ", http.StatusOK, resp.StatusCode)
		return
	}

	req, err = http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the GET request, %s", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Error executing the GET request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("Expected StatusNotImplemented %d, received %d. ", http.StatusNotImplemented, resp.StatusCode)
		return
	}

}

func Test_webhookID(t *testing.T) {
	// instantiate mock HTTP server (just for the purpose of testing
	ts := httptest.NewServer(http.HandlerFunc(webhookID))
	defer ts.Close()

	//create a request to our mock HTTP server
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, ts.URL, nil)
	if err != nil {
		t.Errorf("Error constructing the POST request, %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error executing the POST request, %s", err)
	}

	//check if the response from the handler is what we except
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("Expected StatusNotImplemented %d, received %d. ", http.StatusNotImplemented, resp.StatusCode)
		return
	}
	if resp.StatusCode != 501 {
		t.Errorf("Expected StatusOK %d, received %d. ", 501, resp.StatusCode)
		return
	}

}
