package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
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

	cert, err := tls.LoadX509KeyPair(tc.CertFile, tc.KeyFile)
	if err != nil {
		log.Fatalf("Error loading certificate for aristokicks.ca: %s\n", err)
		return nil, tls.Certificate{}, err
	}
	return reverseProxy, cert, nil
}

func validateFlags(host, targetUrl, certFile, keyFile *string) error {
	if *host == "" {
		return fmt.Errorf("host name invalid")
	}

	if *targetUrl == "" {
		return fmt.Errorf("target url invalid")
	}

	if _, err := os.Stat(*certFile); err != nil {
		return fmt.Errorf("certfile %s invalid: %s", *certFile, err)
	}

	if _, err := os.Stat(*keyFile); err != nil {
		return fmt.Errorf("keyfile %s invalid: %s", *keyFile, err)
	}
	return nil
}

func cliUssageMsg() {
	fmt.Println("Ussage:")
	fmt.Println("\t ./reverse-proxy --host=domain-name.tld \\ #Host domain including port that the reverse proxy will receive request from")
	fmt.Println("\t --targeturl=http://localhost:8080 \\ #The url that the reverse proxy will forward request from the host too")
	fmt.Println("\t --certfile=/path/to/domain/cert.pem \\ #Path to your tls certificate file")
	fmt.Println("\t --keyfile=/path/to/domain/key.pem #Path to your tls key file")
}

type HostHandler struct {
	Host  string
	Proxy *httputil.ReverseProxy
}

func main() {
	var configs []targetConfig

	host := flag.String("host", "", "Host domain including port that the reverse proxy will receive request from: <domain-name.tld>")
	targetUrl := flag.String("targeturl", "", "The url that the reverse proxy will forward request from the host too: <http://localhost:8080>")
	certFile := flag.String("certfile", "", "Path to your tls certificate file: /path/to/domain/cert.pem")
	keyFile := flag.String("keyfile", "", "Path to your tls key file: /path/to/domain/key.pem")

	flag.Parse()

	if err := validateFlags(host, targetUrl, certFile, keyFile); err != nil {
		cliUssageMsg()
		return
	}

	configs = append(configs,
		targetConfig{
			Host:      *host,
			TargetUrl: *targetUrl,
			CertFile:  *certFile,
			KeyFile:   *keyFile,
		})

	for i := 3; i < len(os.Args); i += 4 {
		if os.Args[i] != "" && strings.HasPrefix(os.Args[i], "--host=") &&
			os.Args[i+1] != "" && strings.HasPrefix(os.Args[i+1], "--targeturl=") &&
			os.Args[i+2] != "" && strings.HasPrefix(os.Args[i+2], "--certfile=") &&
			os.Args[i+3] != "" && strings.HasPrefix(os.Args[i+3], "--keyfile=") {
			host = &os.Args[i]
			targetUrl = &os.Args[i+1]
			certFile = &os.Args[i+2]
			keyFile = &os.Args[i+3]

			if err := validateFlags(host, targetUrl, certFile, keyFile); err != nil {
				cliUssageMsg()
				return
			}

			configs = append(configs,
				targetConfig{
					Host:      *host,
					TargetUrl: *targetUrl,
					CertFile:  *certFile,
					KeyFile:   *keyFile,
				})
		}
	}

	var server *http.Server
	var hostHandlers []HostHandler
	var certs []tls.Certificate

	for i := range len(configs) {
		proxy, cert, err := newTarget(&configs[i])
		if err != nil {
			log.Fatalf("Failed to create new reverse proxy target for host %s: %s\n", configs[i].Host, err)
			return
		}
		hostHandlers = append(hostHandlers, HostHandler{Host: configs[i].Host, Proxy: proxy})
		certs = append(certs, cert)
	}

	server = &http.Server{
		Addr: ":443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for i := range len(hostHandlers) {
				if hostHandlers[i].Host != r.Host {
					continue
				} else {
					hostHandlers[i].Proxy.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Not Found", http.StatusNotFound)
		}),
	}

	server.TLSConfig = &tls.Config{
		Certificates: certs,
		MinVersion:   tls.VersionTLS13,
	}

	log.Println("Reverse Proxy listening for connections on :443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Error listening: %s\n", err)
		return
	}
}
