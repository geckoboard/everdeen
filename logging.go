package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

var (
	reqRespLogLine = "Proxy response for request: %s\n**** Request ****\n%s\n**** Response ****\n%s\n\n\n"
)

func logProxyRequestResponse(req *http.Request, resp *http.Response) {
	reqBytes, err := httputil.DumpRequest(req, false)
	if err != nil {
		log.Printf("Error dumping request: %v", req)
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("Error dumping response: %v", req)
	}

	fmt.Printf(reqRespLogLine, req.URL.String(), reqBytes, respBytes)
}

func logExpectationRequest(exp []*Expectation) {
	b, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		log.Printf("Error marshalling expectation request for logging %s", err)
		return
	}

	log.Printf("Register expectations request:\n%s\n", string(b))
}
