package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/j0sh3rs/airprint-proxy/internal"
)

func main() {
	// Command-line flag for the target
	targetFlag := flag.String("target", "", "Target IP or DNS name to proxy AirPrint (IPP) requests")
	portFlag := flag.String("port", "631", "Port to listen on for incoming AirPrint requests")
	debugFlag := flag.Bool("debug", false, "Enable debug logging of all communication to debug.log")
	flag.Parse()

	if *targetFlag == "" {
		fmt.Println("Usage: airprint-proxy -target=<IP or DNS> [-port=<port>] [-debug>]")
		os.Exit(1)
	}

	// Create the proxy server
	proxy, err := internal.NewAirPrintProxy(*targetFlag, *portFlag, *debugFlag)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	// Start the proxy in a separate goroutine
	go func() {
		if err := proxy.Start(); err != nil {
			log.Fatalf("Error starting AirPrint proxy: %v", err)
		}
	}()

	// Listen for OS signals to gracefully shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Stop the server gracefully
	if err := proxy.Shutdown(); err != nil {
		log.Printf("Error shutting down proxy: %v", err)
	}
}
