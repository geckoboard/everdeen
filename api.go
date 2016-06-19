package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/elazarl/goproxy.v1"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
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
	regExp := regexp.MustCompile("/expectations/\\d+/requests")

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

	regExp := regexp.MustCompile("\\d+")
	expId, err := strconv.Atoi(regExp.FindString(r.URL.Path))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	exp := s.findExpectationById(expId)
	if exp == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if found, err := s.requestStore.Where(exp.Id); err != nil {
		http.Error(w, fmt.Sprintf("everdeen: %s", err), http.StatusInternalServerError)
		return
	} else {
		b, _ := json.Marshal(FindResponse{Requests: found})

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
		if expectation.StoreMatchingRequests {
			if expectation.Id == 0 {
				http.Error(w, fmt.Sprintf("everdeen: %s", ExpectationInvalidMsg), StatusUnprocessable)
				log.Printf("ERROR: %v", ExpectationInvalidMsg)
				return
			} else {
				//Check if the expectation is already registered with same id
				if s.findExpectationById(expectation.Id) != nil {
					http.Error(w, fmt.Sprintf("everdeen: %s", ExpectationExistsMsg), StatusUnprocessable)
					log.Printf("ERROR: %v", ExpectationExistsMsg)
					return
				}
			}
		}
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
