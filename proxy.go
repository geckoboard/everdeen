package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
)

func (s *Server) handleProxyRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	expectation, err := s.findMatchingExpectation(r)
	if err != nil {
		return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadGateway, fmt.Sprintf("everdeen: %s", err))
	}

	if expectation == nil {
		return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusNotFound, fmt.Sprintf("everdeen: no expectation matched request"))
	} else {
		expectation.mutex.Lock()
		expectation.Matches += 1
		expectation.mutex.Unlock()

		if expectation.PassThrough {
			return r, nil
		} else {
			return proxyRespond(r, expectation.RespondWith)
		}
	}
}

func (s *Server) findMatchingExpectation(r *http.Request) (*Expectation, error) {
	for _, e := range s.expectations {
		match, err := e.Match(r)
		if err != nil {
			return nil, err
		}

		if match {
			if e.StoreMatchingRequests {
				if err := s.requestStore.Save(e.Uuid, r); err != nil {
					return nil, errors.New(fmt.Sprintf("everdeen: %s", err))
				}
			}
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
