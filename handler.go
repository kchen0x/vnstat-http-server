package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// handleMetrics handles /metrics endpoint, returns Prometheus format metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
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

	// Token authentication is optional for metrics endpoint
	// If token is set, require it; otherwise allow anonymous access
	if s.token != "" && !s.checkToken(r) {
		http.Error(w, "Unauthorized: Invalid or missing token", http.StatusUnauthorized)
		return
	}

	// Get JSON data
	jsonData, err := s.service.GetJSON()
	if err != nil {
		log.Printf("Failed to get JSON data for metrics: %v", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Parse JSON and convert to Prometheus format
	var vnstatData map[string]interface{}
	if err := json.Unmarshal(jsonData, &vnstatData); err != nil {
		log.Printf("Failed to parse JSON data: %v", err)
		http.Error(w, "Failed to parse data", http.StatusInternalServerError)
		return
	}

	// Generate Prometheus metrics
	metrics := s.generatePrometheusMetrics(vnstatData)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metrics))
}

// generatePrometheusMetrics converts vnstat JSON to Prometheus format
func (s *Server) generatePrometheusMetrics(data map[string]interface{}) string {
	var metrics strings.Builder

	// Add help and type comments
	metrics.WriteString("# HELP vnstat_traffic_total_bytes Total traffic in bytes\n")
	metrics.WriteString("# TYPE vnstat_traffic_total_bytes counter\n")
	metrics.WriteString("# HELP vnstat_traffic_month_bytes Monthly traffic in bytes\n")
	metrics.WriteString("# TYPE vnstat_traffic_month_bytes counter\n")
	metrics.WriteString("# HELP vnstat_traffic_today_bytes Today's traffic in bytes\n")
	metrics.WriteString("# TYPE vnstat_traffic_today_bytes counter\n")

	interfaces, ok := data["interfaces"].([]interface{})
	if !ok {
		return metrics.String() + "# No interface data available\n"
	}

	for _, iface := range interfaces {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}

		interfaceName := fmt.Sprintf("%v", ifaceMap["name"])
		// Escape interface name for Prometheus label
		interfaceName = strings.ReplaceAll(interfaceName, "\"", "\\\"")
		interfaceName = strings.ReplaceAll(interfaceName, "\n", "\\n")
		interfaceName = strings.ReplaceAll(interfaceName, "\\", "\\\\")

		traffic, ok := ifaceMap["traffic"].(map[string]interface{})
		if !ok {
			continue
		}

		// Total traffic
		if total, ok := traffic["total"].(map[string]interface{}); ok {
			if rx, ok := total["rx"].(float64); ok {
				metrics.WriteString(fmt.Sprintf("vnstat_traffic_total_bytes{interface=\"%s\",direction=\"rx\"} %.0f\n", interfaceName, rx))
			}
			if tx, ok := total["tx"].(float64); ok {
				metrics.WriteString(fmt.Sprintf("vnstat_traffic_total_bytes{interface=\"%s\",direction=\"tx\"} %.0f\n", interfaceName, tx))
			}
		}

		// Monthly traffic
		if month, ok := traffic["month"].([]interface{}); ok && len(month) > 0 {
			if monthData, ok := month[0].(map[string]interface{}); ok {
				if rx, ok := monthData["rx"].(float64); ok {
					metrics.WriteString(fmt.Sprintf("vnstat_traffic_month_bytes{interface=\"%s\",direction=\"rx\"} %.0f\n", interfaceName, rx))
				}
				if tx, ok := monthData["tx"].(float64); ok {
					metrics.WriteString(fmt.Sprintf("vnstat_traffic_month_bytes{interface=\"%s\",direction=\"tx\"} %.0f\n", interfaceName, tx))
				}
			}
		}

		// Today's traffic (from day array, last element is today)
		if day, ok := traffic["day"].([]interface{}); ok && len(day) > 0 {
			// Get the last element (today's data)
			dayData, ok := day[len(day)-1].(map[string]interface{})
			if ok {
				if rx, ok := dayData["rx"].(float64); ok {
					metrics.WriteString(fmt.Sprintf("vnstat_traffic_today_bytes{interface=\"%s\",direction=\"rx\"} %.0f\n", interfaceName, rx))
				}
				if tx, ok := dayData["tx"].(float64); ok {
					metrics.WriteString(fmt.Sprintf("vnstat_traffic_today_bytes{interface=\"%s\",direction=\"tx\"} %.0f\n", interfaceName, tx))
				}
			}
		}
	}

	return metrics.String()
}

