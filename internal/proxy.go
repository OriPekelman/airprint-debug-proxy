package internal

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// AirPrintProxy holds server settings and references
type AirPrintProxy struct {
	targetURL   string
	port        string
	httpServer  *http.Server
	debugLogger *DebugLogger
}

// NewAirPrintProxy creates a new proxy instance
func NewAirPrintProxy(target, port string, debug bool) (*AirPrintProxy, error) {
	// Initialize debug logger
	debugLogger, err := NewDebugLogger(debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create debug logger: %w", err)
	}

	// Convert the target to a URL
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s", target))
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Create a reverse proxy
	reverseProxy := httputil.NewSingleHostReverseProxy(proxyURL)

	// Customize director to log the request and set correct host header
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		// Save original data for logging
		clientIP, macAddr := getClientInfo(req)
		log.Printf("[INFO] Received AirPrint request from IP: %s, MAC: %s", clientIP, macAddr)

		// Log full request to debug file if enabled
		debugLogger.LogRequest(req, clientIP, macAddr)

		originalDirector(req)
	}

	// Add response logging
	reverseProxy.ModifyResponse = func(resp *http.Response) error {
		return debugLogger.LogResponse(resp)
	}

	mux := http.NewServeMux()
	mux.Handle("/", reverseProxy)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Attempt to proxy the request
			clientIP, _ := getClientInfo(r)
			reverseProxy.ServeHTTP(w, r)

			// If response is written successfully, consider it proxied
			log.Printf("[INFO] Proxied AirPrint request from %s -> %s", clientIP, target)
		}),
	}

	return &AirPrintProxy{
		targetURL:   target,
		port:        port,
		httpServer:  srv,
		debugLogger: debugLogger,
	}, nil
}

// Start begins listening for incoming requests
func (a *AirPrintProxy) Start() error {
	log.Printf("[INFO] Starting AirPrint proxy on port %s -> %s\n", a.port, a.targetURL)
	return a.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (a *AirPrintProxy) Shutdown() error {
	// Close debug logger if active
	if a.debugLogger != nil {
		a.debugLogger.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.httpServer.Shutdown(ctx)
}

// getClientInfo extracts the client IP and attempts to look up its MAC
func getClientInfo(r *http.Request) (ip string, mac string) {
	// We assume RemoteAddr is in the form IP:port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, "unknown"
	}

	// Attempt to get MAC from ARP
	macAddr, err := getMACFromIP(host)
	if err != nil {
		macAddr = "unknown"
	}

	return host, macAddr
}
