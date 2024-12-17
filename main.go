package main

import (
  "fmt"
  "net/http"
  "log"
  "context"
  //"io"
)

// accept http request
// check the host of the request
// send that http request to the correct place

const PORT = 8080

func main() {
  mux := http.NewServeMux()

  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    log.Printf("request host: %s\n", r.Host)
    log.Printf("request URI: %s\n", r.RequestURI)

    var host string 
    switch r.Host {
      case "localhost:8080": {
        host = "0.0.0.0:42069" 
      }
      default: {
        _ = host
        log.Fatalf("we got host a host we have no clue what to do with: %s\n", r.Host)
      }
    }

    req, err := http.NewRequestWithContext(context.Background(), r.Method, fmt.Sprintf("http://%s%s", host, r.RequestURI), r.Body)
    if err != nil {
      log.Fatalf("Error in creating new request: %s\n", err)
    }

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
      log.Fatalf("Error in issueing new request to client: %s\n", err)
    }
    //defer res.Body.Close()

    //body, err := io.ReadAll(res.Body)
    //if err != nil {
    //  log.Fatalf("Failed to read body of response: %s\n", err)
    //}

    //fmt.Fprintf(w, "proxy reponse:\n%s\n", body)
    if err := res.Write(w); err != nil {
      log.Fatalf("Failed to write response from client to http response writer: %s\n", err) 
    } else {
      log.Println("successfully wrote response to client")
    }
  })

  log.Printf("Server listening on port %d\n", PORT) 
  if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), mux); err != nil {
    log.Fatalf("Error listening and serving on port %d: %s\n", PORT, err) 
  } 
}
