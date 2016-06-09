package main

import (
	"flag"
	"log"
	"net/http"

	"gopkg.in/elazarl/goproxy.v1"
)

func main() {
	proxyAddr := flag.String("proxy-addr", ":4321", "Listen address for the HTTP proxy")
	controlAddr := flag.String("control-addr", ":4322", "Listen address for the control API")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()

	server := &Server{Proxy: proxy}
	http.Handle("/", server)
	go http.ListenAndServe(*controlAddr, nil)

	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(server.handleProxyRequest)
	log.Fatal(http.ListenAndServe(*proxyAddr, proxy))
}
