package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

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
	MatchTypeRegex MatchType = "regex"
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
		if criteria.MatchType == "" {
			criteria.MatchType = MatchTypeExact
		}

		switch criteria.MatchType {
		case MatchTypeExact:
			switch criteria.Type {
			case CriteriaTypeMethod:
				conditions = append(conditions, reqMethodMatches(criteria.Value))
			case CriteriaTypeHost:
				conditions = append(conditions, goproxy.ReqHostIs(criteria.Value))
			case CriteriaTypePath:
				conditions = append(conditions, pathIsExactly(criteria.Value))
			case CriteriaTypeHeader:
				conditions = append(conditions, headerIsExactly(criteria.Key, criteria.Value))
			case CriteriaTypeBody:
				conditions = append(conditions, bodyIsExactly(criteria.Value))
			}

		case MatchTypeRegex:
			re, err := regexp.Compile(criteria.Value)
			if err != nil {
				// TODO: Bubble the error up to the HTTP handler
				continue
			}

			switch criteria.Type {
			case CriteriaTypeHost:
				conditions = append(conditions, goproxy.ReqHostMatches(re))
			case CriteriaTypePath:
				conditions = append(conditions, pathMatches(re))
			case CriteriaTypeHeader:
				conditions = append(conditions, headerMatches(criteria.Key, re))
			case CriteriaTypeBody:
				conditions = append(conditions, bodyMatches(re))
			}
		}
	}

	return conditions
}

func headerIsExactly(key, value string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return r.Header.Get(key) == value
	}
}

func headerMatches(key string, re *regexp.Regexp) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return re.MatchString(r.Header.Get(key))
	}
}

func reqMethodMatches(method string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return r.Method == method
	}
}

func pathIsExactly(path string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return r.URL.Path == path
	}
}

func pathMatches(re *regexp.Regexp) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		return re.MatchString(r.URL.Path)
	}
}

func bodyIsExactly(body string) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return false
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

		return string(bodyBytes) == body
	}
}

func bodyMatches(re *regexp.Regexp) goproxy.ReqConditionFunc {
	return func(r *http.Request, _ *goproxy.ProxyCtx) bool {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return false
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

		return re.MatchString(string(bodyBytes))
	}
}

func proxyRespond(rw RespondWith) func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	return func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		resp := &http.Response{}
		resp.Request = r
		resp.TransferEncoding = r.TransferEncoding
		resp.Header = make(http.Header)

		for key, value := range rw.Headers {
			resp.Header.Add(key, value)
		}

		resp.StatusCode = rw.Status

		var bodyReader io.Reader

		if rw.BodyEncoding == BodyEncodingBase64 {
			bodyBytes, err := base64.StdEncoding.DecodeString(rw.Body)

			if err == nil {
				bodyReader = bytes.NewReader(bodyBytes)
			} else {
				resp.StatusCode = http.StatusInternalServerError
				bodyReader = strings.NewReader("everdeen: error decoding base64 encoded body")
			}

			resp.ContentLength = int64(len(bodyBytes))
		} else {
			buf := bytes.NewBufferString(rw.Body)
			resp.ContentLength = int64(buf.Len())
			bodyReader = buf
		}

		resp.Body = ioutil.NopCloser(bodyReader)
		return nil, resp
	}
}
