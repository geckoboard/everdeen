package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gopkg.in/elazarl/goproxy.v1"
)

type testCase struct {
	expectations []Expectation
	scenarios    []scenario
}

type request struct {
	method  string
	url     string
	body    string
	headers map[string]string
}

type response struct {
	status  int
	body    string
	headers map[string]string
}

type scenario struct {
	request  request
	response response
}

func TestMethodExpectation(t *testing.T) {
	websiteServer := buildWebsiteServer()
	defer websiteServer.Close()

	testCases := []testCase{
		// Method Matcher (Exact)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "POST",
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "POST",
						url:    websiteServer.URL,
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Host Matcher
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypeHost,
							Value: "google.com",
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    "http://google.com",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Host Matcher (Regex)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:      CriteriaTypeHost,
							MatchType: MatchTypeRegex,
							Value:     `.*\.google\.com`,
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    "http://images.google.com",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Path Matcher (Exact)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypePath,
							Value: "/foo/bar",
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "/foo/bar",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "/lol",
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Path Matcher (Regex)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:      CriteriaTypePath,
							MatchType: MatchTypeRegex,
							Value:     `/contacts/(\d+)`,
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "/contacts/1",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "/lol",
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Header Matcher (Exact)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypeHeader,
							Key:   "Authorization",
							Value: "Bearer mytoken",
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
						headers: map[string]string{
							"Authorization":    "Bearer mytoken",
							"X-Something-Else": "something else",
						},
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
						headers: map[string]string{
							"Authorization": "something else",
						},
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Header Matcher (Regex)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:      CriteriaTypeHeader,
							Key:       "Authorization",
							MatchType: MatchTypeRegex,
							Value:     `Bearer ([a-z\d]+)`,
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
						headers: map[string]string{
							"Authorization":    "Bearer abc123",
							"X-Something-Else": "something else",
						},
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
						headers: map[string]string{
							"Authorization": "something else",
						},
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Body Matcher (Exact)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypeBody,
							Value: "foo=bar",
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "POST",
						url:    websiteServer.URL,
						body:   "foo=bar",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "POST",
						url:    websiteServer.URL,
						body:   "foo=something-else",
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Body Matcher (Regex)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:      CriteriaTypeBody,
							Value:     "foo=(bar|baz)",
							MatchType: MatchTypeRegex,
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "POST",
						url:    websiteServer.URL,
						body:   "foo=bar",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "POST",
						url:    websiteServer.URL,
						body:   "foo=something-else",
					},
					response{
						status: 200,
						body:   "Got Through",
					},
				},
			},
		},

		// Responding With Custom Headers
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "GET",
						},
					},

					RespondWith{
						Status: 418,
						Body:   "Proxy Response",
						Headers: map[string]string{
							"X-Custom-Header": "Yup it got set",
						},
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
					},
					response{
						status: 418,
						body:   "Proxy Response",
						headers: map[string]string{
							"X-Custom-Header": "Yup it got set",
						},
					},
				},
			},
		},

		// Responding With Binary Body (Base64 Encoded)
		{
			expectations: []Expectation{
				{
					[]Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "GET",
						},
					},

					RespondWith{
						Status:       418,
						Body:         "SGVsbG8gV29ybGQ=",
						BodyEncoding: BodyEncodingBase64,
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL,
					},
					response{
						status: 418,
						body:   "Hello World",
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		runTestCase(t, i, tc)
	}
}

func runTestCase(t *testing.T, i int, tc testCase) {
	proxy, proxyServer, proxyClient := buildProxy()
	defer proxyServer.Close()

	server := Server{Proxy: proxy}

	cer := CreateExpectationsRequest{tc.expectations}
	json, err := json.Marshal(cer)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/expectations", bytes.NewReader(json))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("[%d] unexpected status code while setting up expectations: %d", i, rec.Code)
	}

	for idx, scenario := range tc.scenarios {
		req, err = http.NewRequest(scenario.request.method, scenario.request.url, strings.NewReader(scenario.request.body))
		if err != nil {
			t.Fatalf("[%d - %d] error building request for scenario: %v", i, idx, err)
		}

		for key, value := range scenario.request.headers {
			req.Header.Add(key, value)
		}

		resp, err := proxyClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != scenario.response.status {
			t.Errorf("[%d - %d] unexpected response status, expected: %d, got: %d", i, idx, scenario.response.status, resp.StatusCode)
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("[%d - %d] error reading body: %v", idx, err)
		}

		body := string(bodyBytes)
		if body != scenario.response.body {
			t.Errorf("[%d - %d] unexpected response body, expected: %s, got: %s", i, idx, scenario.response.body, body)
		}

		for key, value := range scenario.response.headers {
			got := resp.Header.Get(key)

			if value != got {
				t.Errorf("[%d - %d] unexpected value for header: %s, expected: %s, got: %s", i, idx, key, value, got)
			}
		}
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
