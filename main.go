package main

import (
	"encoding/json"
	"fmt"
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
		s.Proxy.OnRequest(conditionsForExpectation(expectation)...).DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return nil, goproxy.NewResponse(r, "text/plain", expectation.RespondWith.Status, expectation.RespondWith.Body)
		})
	}
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

func conditionsForExpectation(expectation Expectation) []goproxy.ReqCondition {
	conditions := make([]goproxy.ReqCondition, len(expectation.RequestCriteria))

	for idx, criteria := range expectation.RequestCriteria {
		switch criteria.Type {
		case CriteriaTypeMethod:
			conditions[idx] = func(criteria Criteria) goproxy.ReqConditionFunc {
				return func(r *http.Request, ctx *goproxy.ProxyCtx) bool {
					return r.Method == criteria.Value
				}
			}(criteria)
		}
	}

	return conditions
}
