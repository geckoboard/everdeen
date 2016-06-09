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
	if r.URL.Path == "/ping" {
		fmt.Fprint(w, "PONG")
		return
	}

	if r.URL.Path != "/expectations" {
		http.Error(w, "everdeen: Not Found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		s.listExectations(w, r)
	case "POST":
		s.createExpectations(w, r)
	default:
		http.Error(w, "everdeen: Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listExectations(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if err := json.NewEncoder(w).Encode(s.expectations); err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		log.Printf("ERROR: %v", err)
	}
}

func (s *Server) createExpectations(w http.ResponseWriter, r *http.Request) {
	var request CreateExpectationsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		log.Printf("ERROR: %v", err)
		return
	}

	expectations, err := prepareExpectations(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusBadRequest)
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

		// We expose `Matches` for the `GET /expectations` endpoint
		// but do not want the client to be able to set it.
		e.Matches = 0

		expectations = append(expectations, &e)
	}

	return expectations, nil
}
