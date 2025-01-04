package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSimpleProxy verifies we can create a proxy and simulate a request.
func TestSimpleProxy(t *testing.T) {
	// Create a mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))
	defer backend.Close()

	// Extract the backend server URL (host:port)
	// This is something like: http://127.0.0.1:37753
	// We only need the host portion for the proxy
	proxy, err := NewAirPrintProxy(backend.Listener.Addr().String(), "0")
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	srv := httptest.NewServer(proxy.httpServer.Handler)
	defer srv.Close()

	// Make a request to the proxy
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	if string(body) != "OK" {
		t.Errorf("expected body 'OK', got '%s'", string(body))
	}
}
