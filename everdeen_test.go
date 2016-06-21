package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/satori/go.uuid"
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

var (
	blockedResponse = response{status: 404, body: "everdeen: no expectation matched request"}
)

func TestMethodExpectation(t *testing.T) {
	websiteServer := buildWebsiteServer()
	defer websiteServer.Close()

	testCases := []testCase{
		// Method Matcher (Exact)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "POST",
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Host Matcher
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeHost,
							Value: "google.com",
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Host Matcher (Regex)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:      CriteriaTypeHost,
							MatchType: MatchTypeRegex,
							Value:     `.*\.google\.com`,
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Path Matcher (Exact)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypePath,
							Value: "/foo/bar",
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Path Matcher (Regex)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:      CriteriaTypePath,
							MatchType: MatchTypeRegex,
							Value:     `/contacts/(\d+)`,
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Header Matcher (Exact)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeHeader,
							Key:   "Authorization",
							Value: "Bearer mytoken",
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Header Matcher (Regex)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:      CriteriaTypeHeader,
							Key:       "Authorization",
							MatchType: MatchTypeRegex,
							Value:     `Bearer ([a-z\d]+)`,
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Body Matcher (Exact)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeBody,
							Value: "foo=bar",
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Body Matcher (Regex)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:      CriteriaTypeBody,
							Value:     "foo=(bar|baz)",
							MatchType: MatchTypeRegex,
						},
					},

					RespondWith: RespondWith{
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
					blockedResponse,
				},
			},
		},

		// Query Param Matcher (Single / Exact)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeQueryParam,
							Key:   "q",
							Value: "Search Term",
						},
					},

					RespondWith: RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?q=Search+Term",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?q=Something+Else",
					},
					blockedResponse,
				},
			},
		},

		// Query Param Matcher (Many / Exact)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:   CriteriaTypeQueryParam,
							Key:    "name",
							Values: []string{"Jack", "Sally"},
						},
					},

					RespondWith: RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?name=Sally&name=Jack",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?name=Jack",
					},
					blockedResponse,
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?name=Sally",
					},
					blockedResponse,
				},
			},
		},

		// Query Param Matcher (Regex)
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:      CriteriaTypeQueryParam,
							Key:       "name",
							MatchType: MatchTypeRegex,
							Value:     "Dan(iel)?",
						},
					},

					RespondWith: RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},
				},
			},
			scenarios: []scenario{
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?name=Daniel",
					},
					response{
						status: 418,
						body:   "Proxy Response",
					},
				},
				{
					request{
						method: "GET",
						url:    websiteServer.URL + "?q=Fred",
					},
					blockedResponse,
				},
			},
		},

		// Responding With Custom Headers
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "GET",
						},
					},

					RespondWith: RespondWith{
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
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "GET",
						},
					},

					RespondWith: RespondWith{
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

		// Max Matches
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "GET",
						},
					},

					RespondWith: RespondWith{
						Status: 418,
						Body:   "Proxy Response",
					},

					MaxMatches: 2,
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
					},
				},
				{
					request{
						method: "GET",
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
					blockedResponse,
				},
			},
		},

		// Pass Through
		{
			expectations: []Expectation{
				{
					RequestCriteria: Criteria{
						{
							Type:  CriteriaTypeMethod,
							Value: "GET",
						},
					},

					PassThrough: true,
				},
			},

			scenarios: []scenario{
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

				{
					request{
						method: "POST",
						url:    websiteServer.URL,
					},
					blockedResponse,
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

	server := &Server{Proxy: proxy}
	proxy.OnRequest().DoFunc(server.handleProxyRequest)

	cer := CreateExpectationsRequest{tc.expectations}
	createExpectations(t, server, &cer)

	for idx, scenario := range tc.scenarios {
		req, err := http.NewRequest(scenario.request.method, scenario.request.url, strings.NewReader(scenario.request.body))
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

func TestRequestMatchingExpectationUuid(t *testing.T) {
	proxy, proxyServer, proxyClient := buildProxy()
	defer proxyServer.Close()

	server := &Server{Proxy: proxy}
	proxy.OnRequest().DoFunc(server.handleProxyRequest)

	//Setup an expectation
	expectations := []Expectation{
		{
			StoreMatchingRequests: true,
			RequestCriteria: Criteria{
				{
					Type:  CriteriaTypeMethod,
					Value: "POST",
				},
				{
					Type:  CriteriaTypeHost,
					Value: "www.geckoboard.com",
				},
				{
					Type:  CriteriaTypePath,
					Value: "/hello-world",
				},
			},
			RespondWith: RespondWith{
				Status:       http.StatusOK,
				Body:         "Its me!!!",
				BodyEncoding: BodyEncodingNone,
			},
		},
	}

	cer := CreateExpectationsRequest{expectations}
	exp := createExpectations(t, server, &cer)

	if uuid.Equal(exp[0].Uuid, uuid.Nil) {
		t.Fatalf("Expected returned expectation to have uuid but it didn't: %#v", exp[0])
	}

	req, err := http.NewRequest("POST", "http://www.geckoboard.com/hello-world", strings.NewReader("Some Stuff"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("X-Some-Header", "Hello World")
	req.Header.Add("User-Agent", "Awesome User Agent")

	_, err = proxyClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", "/expectations/"+exp[0].Uuid.String()+"/requests", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status code %d when finding requests with uuid %s", rec.Code, exp[0].Uuid)
	}

	findResponse := FindResponse{}
	if err := json.NewDecoder(rec.Body).Decode(&findResponse); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(findResponse, FindResponse{
		Requests: []Request{
			{
				URL:    "http://www.geckoboard.com/hello-world",
				Method: "POST",
				Headers: map[string][]string{
					"Accept-Encoding": []string{"gzip"},
					"Content-Length":  []string{"10"},
					"X-Some-Header":   []string{"Hello World"},
					"User-Agent":      []string{"Awesome User Agent"},
				},
				BodyBase64: "U29tZSBTdHVmZg==",
			},
		},
	}) {
		t.Errorf("unexpected response from /expectations/%s/requests endpoint: %#v", expectations[0].Uuid, findResponse)
	}
}

func TestRequestsReturnsNotFoundWhenUuidNotExists(t *testing.T) {
	proxy, proxyServer, _ := buildProxy()
	defer proxyServer.Close()

	server := &Server{Proxy: proxy}
	proxy.OnRequest().DoFunc(server.handleProxyRequest)

	req, err := http.NewRequest("GET", "/expectations/1293/requests", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("unexpected status code %d with request id that doesn't exist", rec.Code)
	}
}

func TestRequestReturnsEmptyArrayWithNoMatchingRequests(t *testing.T) {
	proxy, proxyServer, proxyClient := buildProxy()
	defer proxyServer.Close()

	server := &Server{Proxy: proxy}
	proxy.OnRequest().DoFunc(server.handleProxyRequest)

	//Setup an expectation that isn't going to match
	expectations := []Expectation{
		{
			StoreMatchingRequests: true,
			RequestCriteria: Criteria{
				{
					Type:  CriteriaTypeHost,
					Value: "www.google.com",
				},
			},
			RespondWith: RespondWith{
				Status:       http.StatusOK,
				Body:         "Its me!!!",
				BodyEncoding: BodyEncodingNone,
			},
		},
	}

	cer := CreateExpectationsRequest{expectations}
	exp := createExpectations(t, server, &cer)

	if uuid.Equal(exp[0].Uuid, uuid.Nil) {
		t.Fatalf("Expected exp returned to have uuid but it did not %s", exp[0].Uuid)
	}

	req, err := http.NewRequest("GET", "http://www.geckoboard.com/hello-world", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = proxyClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("GET", "/expectations/"+exp[0].Uuid.String()+"/requests", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("unexpected status code %d with expectation uuid %s", rec.Code, exp[0].Uuid)
	}

	if strings.TrimRight(rec.Body.String(), `\n\t`) == `{"requests":[]}` {
		t.Errorf("unexpected response from /expectations/%s/requests endpoint: '%#s'", exp[0].Uuid, rec.Body.String())
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

func createExpectations(t *testing.T, server *Server, cer *CreateExpectationsRequest) []*Expectation {
	data, err := json.Marshal(cer)

	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/expectations", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Unexpected status code [ %d ] while setting up expectations", rec.Code)
	}

	var expectations []*Expectation
	err = json.Unmarshal(rec.Body.Bytes(), &expectations)

	if err != nil {
		t.Fatal(err)
	}

	return expectations
}
