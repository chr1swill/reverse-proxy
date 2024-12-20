package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	// proxy := httputil.NewSingleHostReverseProxy(url)
	// proxy.ModifyResponse = func(resp *http.Response) error {return nil } // if needed
	return httputil.NewSingleHostReverseProxy(url), nil
}

type targetConfig struct {
	Host      string
	TargetUrl string
	CertFile  string
	KeyFile   string
}

func newTarget(tc *targetConfig) (*httputil.ReverseProxy, tls.Certificate, error) {
	reverseProxy, err := newReverseProxy(tc.TargetUrl)
	if err != nil {
		return nil, tls.Certificate{}, err
	}

	// certAristo, err := tls.LoadX509KeyPair("path/to/aristokicks/cert.pem", "path/to/aristokicks/key.pem")
	cert, err := tls.LoadX509KeyPair(tc.CertFile, tc.KeyFile)
	if err != nil {
		log.Fatalf("Error loading certificate for aristokicks.ca: %s\n", err)
		return nil, tls.Certificate{}, err
	}
	return reverseProxy, cert, nil
}

func main() {
	arisokicksTC := &targetConfig{
		Host:      "aristokicks.ca",
		TargetUrl: "http://localhost:8081",
		CertFile:  "path/to/aristokicks/cert.pem",
		KeyFile:   "path/to/aristokicks/key.pem",
	}
	aristokicksProxy, aristokicksCert, err := newTarget(arisokicksTC)
	if err != nil {
		log.Fatalf("Failed to create new reverse proxy target for host %s: %s\n", arisokicksTC.Host, err)
		return
	}

	locciTC := &targetConfig{
		Host:      "locci.ca",
		TargetUrl: "http://localhost:8082",
		CertFile:  "path/to/locci/cert.pem",
		KeyFile:   "path/to/locci/key.pem",
	}

	locciProxy, locciCert, err := newTarget(locciTC)
	if err != nil {
		log.Fatalf("Failed to create new reverse proxy target for host %s: %s\n", arisokicksTC.Host, err)
		return
	}

	server := &http.Server{
		Addr: ":443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Host {
			case arisokicksTC.Host:
				aristokicksProxy.ServeHTTP(w, r)
			case locciTC.Host:
				locciProxy.ServeHTTP(w, r)
			default:
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}),
	}

	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{aristokicksCert, locciCert},
		MinVersion:   tls.VersionTLS13,
	}

	log.Println("Reverse Proxy listening for connections on :443")
	if err = server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Error listening: %s\n", err)
		return
	}
}
