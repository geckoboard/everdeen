package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/geckoboard/everdeen/certs"
	"github.com/satori/go.uuid"
	"github.com/elazarl/goproxy"
)

// Injected by the Makefile
var (
	Version = ""
	GitSHA  = ""
)

var (
	proxyAddr        = flag.String("proxy-addr", ":4321", "Listen address for the HTTP proxy")
	controlAddr      = flag.String("control-addr", ":4322", "Listen address for the control API")
	caCertPath       = flag.String("ca-cert-path", "", "Path to CA certificate file")
	caKeyPath        = flag.String("ca-key-path", "", "Path to CA private key file")
	passthroughMode  = flag.Bool("passthrough-mode", false, "Start up everdeen and default all proxied traffic to passthrough")
	requestBaseStore = flag.String("request-base-store", path.Join(os.TempDir(), "everdeenStore"), "Base store for matching requests")
	generateCA       = flag.Bool("generate-ca-cert", false, "Generate CA certificate and private key for MITM")
)

func main() {
	flag.Parse()

	if *generateCA {
		generateCACert()
	} else {
		//Delete existing path before starting new proxy
		if err := os.RemoveAll(*requestBaseStore); err != nil {
			log.Fatal(err)
		}

		startProxy()
	}
}

func startProxy() {
	fmt.Println("Starting everdeen proxy...")
	fmt.Printf("Version: %s (git sha: %s)\n", Version, GitSHA)
	fmt.Printf("Proxy Address: %s\n", *proxyAddr)
	fmt.Printf("Control Address: %s\n", *controlAddr)
	fmt.Printf("Passthrough all traffic: %t\n", *passthroughMode)

	if *caCertPath != "" && *caKeyPath != "" {
		tlsc, err := tls.LoadX509KeyPair(*caCertPath, *caKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		goproxy.GoproxyCa = tlsc
		fmt.Printf("CA Certificate: %s\n", *caCertPath)
		fmt.Printf("CA Key: %s\n", *caKeyPath)
	} else {
		fmt.Fprintln(os.Stderr, "WARNING: Using default goproxy CA certificate to sign hijacked HTTPS traffic, consider generating your own certificate & key pair!")
	}

	proxy := goproxy.NewProxyHttpServer()

	server := &Server{
		Proxy:        proxy,
		expectations: []*Expectation{},
	}

	if *passthroughMode {
		exp := Expectation{
				Uuid: uuid.NewV4(),
				PassThrough: true,
				RequestCriteria: Criteria{
					{
						Type:  CriteriaTypeHost,
						regexp: regexp.MustCompile(".*"),
						MatchType: MatchTypeRegex,
					},
				},
			}

		server.expectations = append(server.expectations, &exp)
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
