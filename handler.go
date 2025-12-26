package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Server wraps HTTP server configuration
type Server struct {
	token    string
	service  *VnstatService
}

// NewServer creates a new Server instance
func NewServer(token string, service *VnstatService) *Server {
	return &Server{
		token:   token,
		service: service,
	}
}

// addCORS adds CORS response headers
func (s *Server) addCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// checkToken validates the token in the request
func (s *Server) checkToken(r *http.Request) bool {
	// If no token is set, skip authentication
	if s.token == "" {
		return true
	}

	// Get token from query parameters
	token := r.URL.Query().Get("token")
	return token == s.token
}

// handleJSON handles /json endpoint, returns vnstat JSON data
func (s *Server) handleJSON(w http.ResponseWriter, r *http.Request) {
	s.addCORS(w)

	// Handle OPTIONS preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check token authentication
	if !s.checkToken(r) {
		http.Error(w, "Unauthorized: Invalid or missing token", http.StatusUnauthorized)
		return
	}

	// Execute vnstat --json command
	jsonData, err := s.service.GetJSON()
	if err != nil {
		log.Printf("Failed to get JSON data: %v", err)
		// Return JSON formatted error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := map[string]string{
			"error": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Return JSON data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// handleText handles root path / endpoint, returns vnstat text data (monthly view)
func (s *Server) handleText(w http.ResponseWriter, r *http.Request) {
	s.addCORS(w)

	// Handle OPTIONS preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check token authentication
	if !s.checkToken(r) {
		http.Error(w, "Unauthorized: Invalid or missing token", http.StatusUnauthorized)
		return
	}

	// Execute vnstat -m command
	textData, err := s.service.GetText()
	if err != nil {
		log.Printf("Failed to get text data: %v", err)
		// Return plain text error
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v\n", err)
		return
	}

	// Return text data
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(textData)
}

// handleTextGeneric is a generic text handler function
func (s *Server) handleTextGeneric(w http.ResponseWriter, r *http.Request, getData func() ([]byte, error)) {
	s.addCORS(w)

	// Handle OPTIONS preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check token authentication
	if !s.checkToken(r) {
		http.Error(w, "Unauthorized: Invalid or missing token", http.StatusUnauthorized)
		return
	}

	// Execute data retrieval function
	textData, err := getData()
	if err != nil {
		log.Printf("Failed to get data: %v", err)
		// Return plain text error
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v\n", err)
		return
	}

	// Return text data
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(textData)
}

// handleSummary handles /summary endpoint, returns default summary view
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetSummary)
}

// handleDaily handles /daily endpoint, returns daily view
func (s *Server) handleDaily(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetDaily)
}

// handleHourly handles /hourly endpoint, returns hourly view
func (s *Server) handleHourly(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetHourly)
}

// handleWeekly handles /weekly endpoint, returns weekly view
func (s *Server) handleWeekly(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetWeekly)
}

// handleYearly handles /yearly endpoint, returns yearly view
func (s *Server) handleYearly(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetYearly)
}

// handleTop handles /top endpoint, returns top traffic interfaces
func (s *Server) handleTop(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetTop)
}

// handleOneline handles /oneline endpoint, returns one-line output
func (s *Server) handleOneline(w http.ResponseWriter, r *http.Request) {
	s.handleTextGeneric(w, r, s.service.GetOneline)
}

// handleHealth handles /health endpoint for health checks
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.addCORS(w)

	// Handle OPTIONS preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Health check does not require token authentication
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

