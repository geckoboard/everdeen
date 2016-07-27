package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type requestLog struct {
	URL     string
	Method  string
	Headers map[string][]string
	Body    string
}

type responseLog struct {
	Status  int
	Headers map[string][]string
	Body    string
}

func logProxyRequest(r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body from request: %s", err)
		return
	}

	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	rl := &requestLog{
		URL:     r.URL.String(),
		Method:  r.Method,
		Headers: r.Header,
		Body:    string(b),
	}

	rb, err := json.MarshalIndent(rl, "", "  ")
	if err != nil {
		log.Printf("Error marshal request for logging %s", err)
		return
	}

	log.Printf("Request through proxy:\n%s\n", string(rb))
}

func logProxyRequestResponse(r *http.Response) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body from response: %s", err)
		return
	}

	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	rl := &responseLog{
		Status:  r.StatusCode,
		Headers: r.Header,
		Body:    string(b),
	}

	rb, err := json.MarshalIndent(rl, "", "  ")
	if err != nil {
		log.Printf("Error marshaling response for logging %s", err)
		return
	}

	log.Printf("Proxy response for %s\n%s\n", r.Request.URL, string(rb))
}

func logExpectationRequest(exp []*Expectation) {
	b, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		log.Printf("Error marshalling expectation request for logging %s", err)
		return
	}

	log.Printf("Register expectations request:\n%s\n", string(b))
}
