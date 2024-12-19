package main

//import (
//	"context"
//	"fmt"
//	"log"
//	"net/http"
//)
//
//const PORT = 8080
//
//func main() {
//	mux := http.NewServeMux()
//
//	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		log.Printf("request host: %s\n", r.Host)
//		log.Printf("request URI: %s\n", r.RequestURI)
//
//		var host string
//		// modify this so it can be some sort of config or command line args
//		switch r.Host {
//		case "localhost:8080":
//			{
//				host = "0.0.0.0:42069"
//			}
//		default:
//			{
//				_ = host
//				log.Fatalf("we got host a host we have no clue what to do with: %s\n", r.Host)
//			}
//		}
//
//		req, err := http.NewRequestWithContext(context.Background(), r.Method, fmt.Sprintf("http://%s%s", host, r.RequestURI), r.Body)
//		if err != nil {
//			log.Fatalf("Error in creating new request: %s\n", err)
//		}
//
//		client := &http.Client{}
//		res, err := client.Do(req)
//		if err != nil {
//			log.Fatalf("Error in issueing new request to client: %s\n", err)
//		}
//		if err := res.Write(w); err != nil {
//			log.Fatalf("Failed to write response from client to http response writer: %s\n", err)
//		} else {
//			log.Println("successfully wrote response to client")
//		}
//	})
//
//	log.Printf("Server listening on port %d\n", PORT)
//	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), mux); err != nil {
//		log.Fatalf("Error listening and serving on port %d: %s\n", PORT, err)
//	}
//}

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
  //proxy := httputil.NewSingleHostReverseProxy(url)
  //proxy.ModifyResponse = func(resp *http.Response) error {return nil } // if needed
  return httputil.NewSingleHostReverseProxy(url), nil
}

func main() {
  aristoProxy, err := newReverseProxy("http://localhost:8081")
  if err != nil {
    log.Fatalf("Failed to create new reverse proxy for aristokicks.ca: %s\n", err)
    return
  }

  locciProxy, err := newReverseProxy("http://localhost:8082")
  if err != nil {
    log.Fatalf("Failed to create new reverse proxy for aristokicks.ca: %s\n", err)
    return
  }

  server := &http.Server{
    Addr: ":443",
    Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      switch r.Host {
      case "aristokicks.ca":
        aristoProxy.ServeHTTP(w, r)
      case "locci.ca":
        locciProxy.ServeHTTP(w, r)
      default:
        http.Error(w, "Not Found", http.StatusNotFound)
      }
    }),
  }

  certAristo, err := tls.LoadX509KeyPair("path/to/aristokicks/cert.pem", "path/to/aristokicks/key.pem")
  if err != nil {
    log.Fatalf("Error loading certificate for aristokicks.ca: %s\n", err)
    return
  }

  certLocci, err := tls.LoadX509KeyPair("path/to/locci/cert.pem", "path/to/locci/key.pem")
  if err != nil {
    log.Fatalf("Error loading certificate for locci.ca: %s\n", err)
    return
  }

  server.TLSConfig = &tls.Config{
    Certificates: []tls.Certificate{certAristo, certLocci},
    MinVersion: tls.VersionTLS13,
  }

  log.Println("Reverse Proxy listening for connections on :443")
  if err = server.ListenAndServeTLS("", ""); err != nil {
    log.Fatalf("Error listening: %s\n", err)
    return
  }
}
