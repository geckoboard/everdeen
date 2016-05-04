package main

import (
	"log"
	"net/http"

	goproxy "gopkg.in/elazarl/goproxy.v1"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	log.Fatal(http.ListenAndServe(":4321", proxy))
}
