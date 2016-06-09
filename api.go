package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"

	"gopkg.in/elazarl/goproxy.v1"
)

type CreateExpectationsRequest struct {
	Expectations []Expectation `json:"expectations"`
}

type Server struct {
	Proxy *goproxy.ProxyHttpServer

	expectations []*Expectation
	mutex        sync.RWMutex
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request CreateExpectationsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error parsing request payload: %s", err)
		log.Printf("ERROR: %v", err)
		return
	}

	expectations, err := prepareExpectations(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error adding expectations: %s", err)
		log.Printf("ERROR: %v", err)
		return
	}

	s.mutex.Lock()
	for _, expectation := range expectations {
		s.expectations = append(s.expectations, expectation)
	}
	s.mutex.Unlock()
}

func prepareExpectations(request CreateExpectationsRequest) ([]*Expectation, error) {
	expectations := []*Expectation{}

	for _, e := range request.Expectations {
		for _, criterion := range e.RequestCriteria {
			if criterion.MatchType == "" {
				criterion.MatchType = MatchTypeExact
			}

			if criterion.MatchType == MatchTypeRegex {
				var err error
				criterion.regexp, err = regexp.Compile(criterion.Value)

				if err != nil {
					return nil, err
				}
			}
		}

		expectations = append(expectations, &e)
	}

	return expectations, nil
}
