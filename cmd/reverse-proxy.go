package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const PORT = ":443"

func assert(condition bool, message string) {
	if !condition {
		log.Fatalf(message)
	}
}

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
		log.Fatalf("Error loading certificate for %s: %s\n", tc.Host, err)
		return nil, tls.Certificate{}, err
	}
	return reverseProxy, cert, nil
}

func validateTargetSet(host, targetUrl, certFile, keyFile string) error {
	if host == "" {
		return fmt.Errorf("host name invalid: %s", host)
	}

	if targetUrl == "" {
		return fmt.Errorf("target url invalid: %s", targetUrl)
	}

	if _, err := os.Stat(certFile); err != nil {
		return fmt.Errorf("certfile %s invalid: %s", certFile, err)
	}

	if _, err := os.Stat(keyFile); err != nil {
		return fmt.Errorf("keyfile %s invalid: %s", keyFile, err)
	}
	return nil
}

func cliUssageMsg() {
	fmt.Println("")
	fmt.Println("Usage: ./reverse-proxy [target-set] [target-set] ...")
	fmt.Println("")
	fmt.Println("\t[target-set] := <HOST> <TARGETURL> <CERTFILE> <KEYFILE>")
	fmt.Println("")
	fmt.Println("\t--host      : Domain the reverse proxy will receive request on behave of")
	fmt.Println("\t--targeturl : The url reverse proxy will forward the request too")
	fmt.Println("\t--certfile  : Path to your tls certificate file")
	fmt.Println("\t--keyfile   : Path to your tls private key file")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("\t./reverse-proxy \\")
	fmt.Println("\t--host=domain.tld \\")
	fmt.Println("\t--targeturl=http://localhost:8080 \\")
	fmt.Println("\t--certfile=/path/to/domain/cert.pem \\")
	fmt.Println("\t--keyfile=/path/to/domain/privkey.pem")
	fmt.Println("")
}

type HostHandler struct {
	Host  string
	Proxy *httputil.ReverseProxy
}

const ARGS_IN_TARGET_SET = 4

func collectTargetSets() [][ARGS_IN_TARGET_SET]string {
	collectedTargetSets := make([][ARGS_IN_TARGET_SET]string, 0, (len(os.Args)-1)/ARGS_IN_TARGET_SET)

	assert((len(os.Args)-1)%4 == 0,
		fmt.Sprintf("Not able to process input, length of argurment is not divisible by %d\n",
			ARGS_IN_TARGET_SET))

	args := os.Args[1:]

	for i := 0; i < len(args)/ARGS_IN_TARGET_SET; i++ {
		var targetSet [ARGS_IN_TARGET_SET]string
		for j := range ARGS_IN_TARGET_SET {
			targetSet[j] = args[(i*ARGS_IN_TARGET_SET)+j]
		}
		collectedTargetSets = append(collectedTargetSets, targetSet)
	}

	return collectedTargetSets
}

func parseArgsToTargetSets() ([][4]string, error) {
	numberOfArgs := len(os.Args) - 1
	if numberOfArgs%ARGS_IN_TARGET_SET != 0 {
		return nil, fmt.Errorf("number of args should be divisible by four but there was a remainder of %d", numberOfArgs%4)
	}

	collectionOfTargetSets := collectTargetSets()
	for i := range len(collectionOfTargetSets) {
		foundHost, foundTargetUrl, foundCertFile, foundKeyFile := false, false, false, false
    fmt.Printf("current targetSet: %v\n", collectionOfTargetSets[i])

		for j := range ARGS_IN_TARGET_SET {
			currentArg := collectionOfTargetSets[i][j]
			if strings.HasPrefix(currentArg, "--host=") {
				foundHost = true
			} else if strings.HasPrefix(currentArg, "--targeturl=") {
				foundTargetUrl = true
			} else if strings.HasPrefix(currentArg, "--certfile=") {
				foundCertFile = true
			} else if strings.HasPrefix(currentArg, "--keyfile=") {
				foundKeyFile = true
			}
		}

		if !foundHost {
			return nil, fmt.Errorf("syntax error parsing target-set at index: %d missing --host=<?> arg", i)
		}

		if !foundTargetUrl {
			return nil, fmt.Errorf("syntax error parsing target-set at index: %d missing --targeturl=<?> arg", i)
		}

		if !foundCertFile {
			return nil, fmt.Errorf("syntax error parsing target-set at index: %d missing --certfile=<?> arg", i)
		}

		if !foundKeyFile {
			return nil, fmt.Errorf("syntax error parsing target-set at index: %d missing --keyfile=<?> arg", i)
		}
	}

	return collectionOfTargetSets, nil
}

func toTargetConfig(collectionOfTargetSets [][ARGS_IN_TARGET_SET]string) []targetConfig {
	collectionOfTargetConfigs := make([]targetConfig, 0, len(collectionOfTargetSets))

	for i := range len(collectionOfTargetSets) {
		tc := &targetConfig{}

		for j := range ARGS_IN_TARGET_SET {
			currentSetMember := collectionOfTargetSets[i][j]

			if strings.HasPrefix(currentSetMember, "--host=") {
				tc.Host = strings.TrimPrefix(currentSetMember, "--host=")
			} else if strings.HasPrefix(currentSetMember, "--targeturl=") {
				tc.TargetUrl = strings.TrimPrefix(currentSetMember, "--targeturl=")
			} else if strings.HasPrefix(currentSetMember, "--certfile=") {
				tc.CertFile = strings.TrimPrefix(currentSetMember, "--certfile=")
			} else if strings.HasPrefix(currentSetMember, "--keyfile=") {
				tc.KeyFile = strings.TrimPrefix(currentSetMember, "--keyfile=")
			}
		}

		collectionOfTargetConfigs = append(collectionOfTargetConfigs, *tc)
	}
	return collectionOfTargetConfigs
}

func main() {
	if (len(os.Args)-1)%4 != 0 {
		log.Printf("Malformated args, all [target-sets] must contain <HOST> <TARGETURL> <CERTFILE> <KEYFILE>\n")
	}

	targetSets, err := parseArgsToTargetSets()
	if err != nil {
		log.Printf("Error with target-sets: %s\n", err)
	}

	var server *http.Server
	var hostHandlers []HostHandler
	var certs []tls.Certificate
	configs := toTargetConfig(targetSets)

	for i := range len(configs) {
		if err := validateTargetSet(configs[i].Host,
			configs[i].TargetUrl, configs[i].CertFile,
			configs[i].KeyFile); err != nil {
			log.Printf("Error with target-set %d: %s\n\n", i, err)
			cliUssageMsg()
			return
		}

		proxy, cert, err := newTarget(&configs[i])
		if err != nil {
			log.Fatalf("Failed to create new reverse proxy target for host %s: %s\n", configs[i].Host, err)
			return
		}
		hostHandlers = append(hostHandlers, HostHandler{Host: configs[i].Host, Proxy: proxy})
		certs = append(certs, cert)
	}

	server = &http.Server{
		Addr: PORT,
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

	log.Println("Reverse Proxy listening for connections on %s", PORT)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Error listening: %s\n", err)
		return
	}
}
