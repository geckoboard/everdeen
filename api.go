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

	s.mutex.Lock()
	for _, e := range request.Expectations {
		// TODO: Move this somewhere else?
		// Precompile Regexps
		for _, criterion := range e.RequestCriteria {
			if criterion.MatchType == "" {
				criterion.MatchType = MatchTypeExact
			}

			if criterion.MatchType == MatchTypeRegex {
				var err error
				criterion.regexp, err = regexp.Compile(criterion.Value)

				// TODO: Handle this error properly
				if err != nil {
					panic(err)
				}
			}
		}

		s.expectations = append(s.expectations, &e)
	}
	s.mutex.Unlock()
}
