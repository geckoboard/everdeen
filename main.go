package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/geckoboard/everdeen/certs"

	"gopkg.in/elazarl/goproxy.v1"
)

var (
	proxyAddr     = flag.String("proxy-addr", ":4321", "Listen address for the HTTP proxy")
	controlAddr   = flag.String("control-addr", ":4322", "Listen address for the control API")
	caCertPath    = flag.String("ca-cert-path", "", "Path to CA certificate file")
	caKeyPath     = flag.String("ca-key-path", "", "Path to CA private key file")
	storeRequests = flag.Bool("store-requests", false, "TODO: description")

	generateCA = flag.Bool("generate-ca-cert", false, "Generate CA certificate and private key for MITM")
)

func main() {
	flag.Parse()

	if *generateCA {
		generateCACert()
	} else {
		startProxy()
	}
}

func startProxy() {
	if *caCertPath != "" && *caKeyPath != "" {
		tlsc, err := tls.LoadX509KeyPair(*caCertPath, *caKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		goproxy.GoproxyCa = tlsc
	}

	proxy := goproxy.NewProxyHttpServer()

	server := &Server{
		Proxy:         proxy,
		expectations:  []*Expectation{},
		storeRequests: *storeRequests,
	}
	http.Handle("/", server)
	go http.ListenAndServe(*controlAddr, nil)

	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(server.handleProxyRequest)
	log.Fatal(http.ListenAndServe(*proxyAddr, proxy))
}

func generateCACert() {
	_, err := certs.GenerateAndSave("everdeen.proxy", "Everdeen Authority", 365*24*time.Hour)
	if err != nil {
		log.Fatal(err)
	}
}
