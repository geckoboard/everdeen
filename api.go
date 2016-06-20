package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/satori/go.uuid"
	"gopkg.in/elazarl/goproxy.v1"
)

type CreateExpectationsRequest struct {
	Expectations []Expectation `json:"expectations"`
}

type FindResponse struct {
	Requests []Request `json:"requests"`
}

type Server struct {
	Proxy *goproxy.ProxyHttpServer

	expectations []*Expectation
	mutex        sync.RWMutex
	requestStore RequestStore
}

const (
	StatusUnprocessable   int    = 422
	ExpectationInvalidMsg string = "Expectation requires an id when used with store requests"
	ExpectationExistsMsg  string = "Expectation with that id already exists"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	regExp := regexp.MustCompile("/expectations/[a-f0-9-]+/requests")

	if r.Method == "GET" && regExp.MatchString(r.URL.Path) {
		s.findRequests(w, r)
		return
	}

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

	s.requestStore.mutex.RLock()
	defer s.requestStore.mutex.RUnlock()

	expUuid, err := uuid.FromString(strings.Split(r.URL.Path, "/")[2])

	if uuid.Equal(expUuid, uuid.Nil) || err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	exp := s.findExpectationByUuid(expUuid)
	if exp == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if found, err := s.requestStore.Where(exp.Uuid); err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		return
	} else {
		b, err := json.Marshal(FindResponse{Requests: found})

		if err != nil {
			http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", b)
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
		//User shouldn't be setting and is handled by server
		expectation.Uuid = uuid.NewV4()
		s.expectations = append(s.expectations, expectation)
	}

	s.mutex.Unlock()

	data, err := json.Marshal(expectations)

	if err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		log.Printf("ERROR: %v", err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
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
