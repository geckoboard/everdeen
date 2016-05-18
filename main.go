package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"gopkg.in/elazarl/goproxy.v1"
)

type CreateExpectationsRequest struct {
	Expectations []Expectation `json:"expectations"`
}

type Expectation struct {
	RequestCriteria []Criteria  `json:"request_criteria"`
	RespondWith     RespondWith `json:"respond_with"`
}

type CriteriaType string

const (
	CriteriaTypeMethod CriteriaType = "method"
	CriteriaTypeHost   CriteriaType = "host"
	CriteriaTypePath   CriteriaType = "path"
	CriteriaTypeHeader CriteriaType = "header"
	CriteriaTypeBody   CriteriaType = "body"
)

type MatchType string

const (
	MatchTypeExact MatchType = "exact"
)

type BodyEncoding string

const (
	BodyEncodingNone   BodyEncoding = ""
	BodyEncodingBase64 BodyEncoding = "base64"
)

type Criteria struct {
	Type      CriteriaType `json:"type"`
	Key       string       `json:"key"`
	MatchType MatchType    `json:"match_type"`
	Value     string       `json:"value"`
}

type RespondWith struct {
	Status       int               `json:"status"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	BodyEncoding BodyEncoding      `json:"body_encoding"`
}

func main() {
	proxy := goproxy.NewProxyHttpServer()

	server := Server{Proxy: proxy}
	http.Handle("/expectations", server)
	go http.ListenAndServe(":4322", nil)

	proxy.Verbose = true
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)
	log.Fatal(http.ListenAndServe(":4321", proxy))
}

/*

POST /expectations

TODO: Query String?
TODO: Times
TODO: How do we stop unmocked external requests?

{
	"expectations": [
		{
			"request_criteria": [
				{ "type": "method", "match_type": "exact", "value": "GET" },
				{ "type": "header", "key": "Host", "match_type": "regexp", "value": "^geckoboard$" },
				{ "type": "header", "key": "User-Agent", "match_type": "regexp", "value": "Chrome" },
				{ "type": "body", "match_type": "regexp", "value": "^geckoboard$" },
				{ "type": "url", "match_type": "regexp", "value": "^geckoboard$" }
			],

			"respond_with": {
				"status": 200,
				"headers": {
					"X-Request-Id": "abc123"
				},
				"body": "abc123",
				"body_encoding": "base64"
			}
		}
	]
}

*/

type Server struct {
	Proxy *goproxy.ProxyHttpServer
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO? Test this?
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Are you drunk?")
		return
	}

	var request CreateExpectationsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Something went boom!")
		log.Printf("ERROR: %v", err)
		return
	}

	for _, expectation := range request.Expectations {
		conditions := conditionsForExpectation(expectation)

		s.Proxy.OnRequest(conditions...).
			DoFunc(proxyRespond(expectation.RespondWith))
	}
}

func conditionsForExpectation(expectation Expectation) []goproxy.ReqCondition {
	conditions := []goproxy.ReqCondition{}

	for _, criteria := range expectation.RequestCriteria {
		switch criteria.Type {
		case CriteriaTypeMethod:
			conditions = append(conditions, reqMethodMatches(criteria.Value))
		case CriteriaTypeHost:
			conditions = append(conditions, goproxy.ReqHostIs(criteria.Value))
		case CriteriaTypePath:
			conditions = append(conditions, pathMatches(criteria.Value))
		case CriteriaTypeHeader:
			conditions = append(conditions, headerMatches(criteria.Key, criteria.Value))
		case CriteriaTypeBody:
			conditions = append(conditions, bodyMatches(criteria.Value))
		}
	}

	return conditions
}

func headerMatches(key, value string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return r.Header.Get(key) == value
	}
}

func reqMethodMatches(method string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return r.Method == method
	}
}

func pathMatches(path string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return r.URL.Path == path
	}
}

func bodyMatches(body string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return false
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

		return string(bodyBytes) == body
	}
}

func proxyRespond(rw RespondWith) func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	return func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return nil, goproxy.NewResponse(r, "text/plain", rw.Status, rw.Body)
	}
}
