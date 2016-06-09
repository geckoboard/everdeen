package main

import (
	"log"
	"net/http"

	"gopkg.in/elazarl/goproxy.v1"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()

	server := &Server{Proxy: proxy}
	http.Handle("/expectations", server)
	go http.ListenAndServe(":4322", nil)

	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(server.handleProxyRequest)
	log.Fatal(http.ListenAndServe(":4321", proxy))
}
