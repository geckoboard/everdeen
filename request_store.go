package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
)

type storedRequest struct {
	urlString string
	method    string
	headers   map[string][]string
	bodyBytes []byte
}

type RequestStore struct {
	requests []storedRequest
	mutex    sync.RWMutex
}

func (rs *RequestStore) Save(r *http.Request) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	var b []byte
	var err error

	if r.Body != nil {
		defer r.Body.Close()

		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
	}

	request := storedRequest{
		urlString: r.URL.String(),
		method:    r.Method,
		headers:   r.Header,
		bodyBytes: b,
	}

	rs.requests = append(rs.requests, request)
	return nil
}

func (rs *RequestStore) Where(predicate func(*http.Request) (bool, error)) ([]*http.Request, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	found := []*http.Request{}

	for _, storedRequest := range rs.requests {
		req, err := http.NewRequest(storedRequest.method, storedRequest.urlString, bytes.NewReader(storedRequest.bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header = storedRequest.headers

		matched, err := predicate(req)
		if err != nil {
			return nil, err
		}

		if matched {
			found = append(found, req)
		}
	}

	return found, nil
}
