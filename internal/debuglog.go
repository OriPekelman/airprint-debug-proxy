package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"
)

// DebugLogger handles logging of all HTTP communication to a file
type DebugLogger struct {
	file   *os.File
	mu     sync.Mutex
	active bool
}

// NewDebugLogger creates a new debug logger that writes to the specified file
func NewDebugLogger(enabled bool) (*DebugLogger, error) {
	if !enabled {
		return &DebugLogger{active: false}, nil
	}

	filename := fmt.Sprintf("airprint-debug-%s.log", time.Now().Format("2006-01-02-15-04-05"))
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create debug log file: %w", err)
	}

	log.Printf("[INFO] Debug logging enabled. Writing to: %s", filename)

	return &DebugLogger{
		file:   file,
		active: true,
	}, nil
}

// LogRequest logs the full HTTP request including headers and body
func (d *DebugLogger) LogRequest(r *http.Request, clientIP, macAddr string) {
	if !d.active {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Write separator and timestamp
	separator := "\n" + strings.Repeat("=", 80) + "\n"
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	d.file.WriteString(separator)
	d.file.WriteString(fmt.Sprintf("REQUEST at %s\n", timestamp))
	d.file.WriteString(fmt.Sprintf("Client IP: %s | MAC: %s\n", clientIP, macAddr))
	d.file.WriteString(strings.Repeat("=", 80) + "\n")

	// Dump the full request
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		d.file.WriteString(fmt.Sprintf("Error dumping request: %v\n", err))
		return
	}

	d.file.Write(dump)
	d.file.WriteString("\n")
	d.file.Sync()
}

// LogResponse logs the full HTTP response including headers and body
func (d *DebugLogger) LogResponse(resp *http.Response) error {
	if !d.active {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	d.file.WriteString(fmt.Sprintf("\nRESPONSE at %s\n", timestamp))
	d.file.WriteString(strings.Repeat("-", 80) + "\n")

	// Read the body so we can log it
	var bodyBytes []byte
	if resp.Body != nil {
		bodyBytes, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	// Create a new response with the body for dumping
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		d.file.WriteString(fmt.Sprintf("Error dumping response: %v\n", err))
		// Restore the body even if dump failed
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return nil
	}

	d.file.Write(dump)
	d.file.WriteString("\n")

	// Restore the body for the actual response
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	d.file.Sync()
	return nil
}

// Close closes the debug log file
func (d *DebugLogger) Close() error {
	if !d.active || d.file == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.file.Close()
}
