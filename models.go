package main

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/satori/go.uuid"
)

type CriteriaType string

const (
	CriteriaTypeMethod     CriteriaType = "method"
	CriteriaTypeHost       CriteriaType = "host"
	CriteriaTypePath       CriteriaType = "path"
	CriteriaTypeHeader     CriteriaType = "header"
	CriteriaTypeBody       CriteriaType = "body"
	CriteriaTypeQueryParam CriteriaType = "query_param"
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

type Request struct {
	URL        string              `json:"url"`
	Method     string              `json:"method"`
	Headers    map[string][]string `json:"headers"`
	BodyBase64 string              `json:"body_base64"`
}

type Expectation struct {
	RequestCriteria       Criteria    `json:"request_criteria"`
	RespondWith           RespondWith `json:"respond_with"`
	MaxMatches            int         `json:"max_matches"`
	PassThrough           bool        `json:"pass_through"`
	StoreMatchingRequests bool        `json:"store_matching_requests"`
	Uuid                  uuid.UUID   `json:"uuid"`

	Matches int `json:"matches"`
	mutex   sync.RWMutex
}

func (e *Expectation) Match(r *http.Request) (bool, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.MaxMatches > 0 && e.Matches >= e.MaxMatches {
		return false, nil
	}

	return e.RequestCriteria.Match(r)
}

type Criteria []*Criterion

func (c Criteria) Match(r *http.Request) (bool, error) {
	for _, criterion := range c {
		match, err := criterion.Match(r)

		if err != nil || !match {
			return false, err
		}
	}

	return true, nil
}

type Criterion struct {
	Type      CriteriaType `json:"type"`
	Key       string       `json:"key"`
	MatchType MatchType    `json:"match_type"`
	Value     string       `json:"value"`
	Values    []string     `json:"values"`

	regexp *regexp.Regexp
}

func (c *Criterion) Match(r *http.Request) (bool, error) {
	switch c.MatchType {
	case MatchTypeExact:
		switch c.Type {
		case CriteriaTypeMethod:
			return reqMethodIsExactly(r, c.Value)
		case CriteriaTypeHost:
			return reqHostIsExactly(r, c.Value)
		case CriteriaTypePath:
			return pathIsExactly(r, c.Value)
		case CriteriaTypeHeader:
			return headerIsExactly(r, c.Key, c.Value)
		case CriteriaTypeBody:
			return bodyIsExactly(r, c.Value)
		case CriteriaTypeQueryParam:
			if len(c.Values) == 0 {
				return queryParamIsExactly(r, c.Key, c.Value)
			} else {
				return queryParamIsAllOf(r, c.Key, c.Values)
			}
		}

	case MatchTypeRegex:
		switch c.Type {
		case CriteriaTypeHost:
			return reqHostMatches(r, c.regexp)
		case CriteriaTypePath:
			return pathMatches(r, c.regexp)
		case CriteriaTypeHeader:
			return headerMatches(r, c.Key, c.regexp)
		case CriteriaTypeBody:
			return bodyMatches(r, c.regexp)
		case CriteriaTypeQueryParam:
			return queryParamMatches(r, c.Key, c.regexp)
		}
	}

	return true, nil
}

type RespondWith struct {
	Status       int               `json:"status"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	BodyEncoding BodyEncoding      `json:"body_encoding"`
}
