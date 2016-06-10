package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sync"

	"gopkg.in/elazarl/goproxy.v1"
)

type CreateExpectationsRequest struct {
	Expectations []Expectation `json:"expectations"`
}

type FindRequest struct {
	RequestCriteria Criteria `json:"request_criteria"`
}

type FindResponse struct {
	Requests []Request `json:"requests"`
}

type Server struct {
	Proxy *goproxy.ProxyHttpServer

	expectations []*Expectation
	mutex        sync.RWMutex

	storeRequests bool
	requests      RequestStore
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ping":
		fmt.Fprint(w, "PONG")
		return
	case "/expectations":
		switch r.Method {
		case "GET":
			s.listExectations(w, r)
		case "POST":
			s.createExpectations(w, r)
		default:
			http.Error(w, "everdeen: Method Not Allowed", http.StatusMethodNotAllowed)
		}
	case "/requests":
		if r.Method != "POST" {
			http.Error(w, "everdeen: Method Not Allowed", http.StatusMethodNotAllowed)
		}

		s.findRequests(w, r)
	default:
		http.Error(w, "everdeen: Not Found", http.StatusNotFound)
	}
}

func (s *Server) findRequests(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.requests.mutex.RLock()
	defer s.requests.mutex.RUnlock()

	var request FindRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		log.Printf("ERROR: %v", err)
		return
	}

	prepareCriteria(request.RequestCriteria)
	found, err := s.requests.Where(request.RequestCriteria.Match)
	if err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		log.Printf("ERROR: %v", err)
		return
	}

	response := FindResponse{
		Requests: make([]Request, len(found)),
	}

	for idx, req := range found {
		bytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
			log.Printf("ERROR: %v", err)
			return
		}
		bodyBase64 := base64.StdEncoding.EncodeToString(bytes)

		response.Requests[idx] = Request{
			URL:        req.URL.String(),
			Method:     req.Method,
			Headers:    req.Header,
			BodyBase64: bodyBase64,
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		log.Printf("ERROR: %v", err)
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
		prepareCriteria(e.RequestCriteria)

		// We expose `Matches` for the `GET /expectations` endpoint
		// but do not want the client to be able to set it.
		e.Matches = 0

		expectations = append(expectations, &e)
	}

	return expectations, nil
}

func prepareCriteria(c Criteria) error {
	for _, criterion := range c {
		if criterion.MatchType == "" {
			criterion.MatchType = MatchTypeExact
		}

		if criterion.MatchType == MatchTypeRegex {
			var err error
			criterion.regexp, err = regexp.Compile(criterion.Value)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
