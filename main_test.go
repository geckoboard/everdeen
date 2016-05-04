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
	proxy, proxyServer, proxyClient := buildProxy()
	defer proxyServer.Close()

	websiteServer := buildWebsiteServer()
	defer websiteServer.Close()

	server := Server{Proxy: proxy}

	createExpectations(t, server,
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
						"status": 418,
						"body": "Proxy Response"
					}
				}
			]
		}`,
	)

	for _, example := range []struct {
		method string
		status int
		body   string
	}{
		{"POST", http.StatusTeapot, "Proxy Response"},
		{"GET", http.StatusOK, "Got Through"},
		{"PUT", http.StatusOK, "Got Through"},
		{"DELETE", http.StatusOK, "Got Through"},
	} {
		req, err := http.NewRequest(example.method, websiteServer.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := proxyClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		checkResponse(t, resp, example.status, example.body)
	}
}

func buildProxy() (*goproxy.ProxyHttpServer, *httptest.Server, *http.Client) {
	proxy := goproxy.NewProxyHttpServer()
	proxyServer := httptest.NewServer(proxy)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(proxyServer.URL)
			},
		},
	}

	return proxy, proxyServer, client
}

func buildWebsiteServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Got Through")
	}))
}

func createExpectations(t *testing.T, server Server, json string) {
	req := buildExpectationsRequest(t, json)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected response code: %d", rec.Code)
	}
}

func buildExpectationsRequest(t *testing.T, json string) *http.Request {
	req, err := http.NewRequest("POST", "/expectations", strings.NewReader(json))
	if err != nil {
		t.Fatal(err)
	}

	return req
}

func checkResponse(t *testing.T, resp *http.Response, status int, body string) {
	if resp.StatusCode != status {
		t.Errorf("unexpected response code: %d", status)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	str := string(bytes)
	if body != str {
		t.Errorf("unexpected response body: %s", str)
	}
}
