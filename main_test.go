package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gopkg.in/elazarl/goproxy.v1"
)

func TestMethodExpectation(t *testing.T) {
	proxy := goproxy.NewProxyHttpServer()

	proxyServer := httptest.NewServer(proxy)
	defer proxyServer.Close()

	websiteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Got Through!")
	}))
	defer websiteServer.Close()

	server := Server{Proxy: proxy}

	req, err := http.NewRequest("POST", "/expectations", strings.NewReader(
		`{
			"expectations": [
				{
					"request_criteria": [
						{
							"type": "method",
							"value": "POST"
						}
					],

					"respond_with": {
						"status": 200,
						"body": "Hello World"
					}
				}
			]
		}`,
	))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected response code: %d", rec.Code)
	}

	req, err = http.NewRequest("POST", websiteServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(proxyServer.URL)
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response code (via proxy): %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	body := string(bytes)
	if body != "Hello World" {
		t.Errorf("unexpected response body: %s", body)
	}

	req, err = http.NewRequest("GET", websiteServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = string(bytes)
	if body != "Got Through!" {
		t.Errorf("unexpected response body: %s", body)
	}
}
