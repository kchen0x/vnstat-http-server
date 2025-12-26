package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Parse command line arguments
	port := flag.String("port", "8080", "Listening port")
	token := flag.String("token", "", "Authentication token (leave empty to disable)")
	interfaceName := flag.String("interface", "", "Network interface name (leave empty to query all)")
	flag.Parse()

	// Create VnstatService instance
	service := NewVnstatService(*interfaceName)

	// Check if vnstat is installed before starting
	if err := service.CheckVnstatInstalled(); err != nil {
		log.Fatalf("Failed to start: %v\nPlease ensure vnstat is installed", err)
	}

	// Create Server instance
	server := NewServer(*token, service)

	// Register routes (specific paths must be registered before generic paths)
	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/json", server.handleJSON)
	http.HandleFunc("/summary", server.handleSummary)
	http.HandleFunc("/daily", server.handleDaily)
	http.HandleFunc("/hourly", server.handleHourly)
	http.HandleFunc("/weekly", server.handleWeekly)
	http.HandleFunc("/yearly", server.handleYearly)
	http.HandleFunc("/top", server.handleTop)
	http.HandleFunc("/oneline", server.handleOneline)
	http.HandleFunc("/", server.handleText) // Default monthly view

	// Print startup information
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("vnstat-http-server started successfully")
	log.Printf("Listening on: http://0.0.0.0%s", addr)
	if *token != "" {
		log.Printf("Token authentication: enabled")
		log.Printf("Example: http://localhost%s/json?token=%s", addr, *token)
	} else {
		log.Printf("Token authentication: disabled (recommended to enable in production)")
		log.Printf("Example: http://localhost%s/json", addr)
	}
	log.Printf("Health check: http://localhost%s/health", addr)
	log.Printf("Available endpoints: /json, /summary, /daily, /hourly, /weekly, /monthly(/), /yearly, /top, /oneline")
	log.Printf("Press Ctrl+C to stop")

	// Start HTTP server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}

