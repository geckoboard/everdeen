package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
)

func reqMethodIsExactly(r *http.Request, method string) (bool, error) {
	return r.Method == method, nil
}

func reqHostIsExactly(r *http.Request, host string) (bool, error) {
	return r.URL.Host == host, nil
}

func pathIsExactly(r *http.Request, path string) (bool, error) {
	return r.URL.Path == path, nil
}

func headerIsExactly(r *http.Request, key, value string) (bool, error) {
	return r.Header.Get(key) == value, nil
}

func bodyIsExactly(r *http.Request, body string) (bool, error) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return false, err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	return string(bodyBytes) == body, nil
}

func reqHostMatches(r *http.Request, re *regexp.Regexp) (bool, error) {
	return re.MatchString(r.URL.Host), nil
}

func pathMatches(r *http.Request, re *regexp.Regexp) (bool, error) {
	return re.MatchString(r.URL.Path), nil
}

func headerMatches(r *http.Request, key string, re *regexp.Regexp) (bool, error) {
	return re.MatchString(r.Header.Get(key)), nil
}

func bodyMatches(r *http.Request, re *regexp.Regexp) (bool, error) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return false, err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	return re.MatchString(string(bodyBytes)), nil
}
