package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/elazarl/goproxy.v1"
)

func (s *Server) handleProxyRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	expectation, err := s.findMatchingExpectation(r)
	if err != nil {
		return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadGateway, fmt.Sprintf("everdeen: %s", err))
	}

	if expectation != nil {
		expectation.mutex.Lock()
		expectation.matches += 1
		expectation.mutex.Unlock()
		return proxyRespond(r, expectation.RespondWith)
	}

	return r, nil
}

func (s *Server) findMatchingExpectation(r *http.Request) (*Expectation, error) {
	for _, e := range s.expectations {
		match, err := e.Match(r)
		if err != nil {
			return nil, err
		}

		if match {
			return e, nil
		}
	}

	return nil, nil
}

func proxyRespond(r *http.Request, rw RespondWith) (*http.Request, *http.Response) {
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
