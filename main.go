package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

func main() {
	// Parse command line arguments
	port := flag.String("port", "8080", "Listening port")
	token := flag.String("token", "", "Authentication token (leave empty to disable)")
	interfaceName := flag.String("interface", "", "Network interface name (leave empty to query all)")
	
	// Grafana Cloud push configuration
	grafanaURL := flag.String("grafana-url", "", "Grafana Cloud Prometheus remote write URL (e.g., https://YOUR_PROMETHEUS_INSTANCE.grafana.net/api/prom/push)")
	grafanaUser := flag.String("grafana-user", "", "Grafana Cloud instance ID")
	grafanaToken := flag.String("grafana-token", "", "Grafana Cloud API token")
	grafanaInterval := flag.Duration("grafana-interval", 30*time.Second, "Interval for pushing metrics to Grafana Cloud")
	
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
	http.HandleFunc("/metrics", server.handleMetrics)
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
	log.Printf("Available endpoints: /json, /metrics, /summary, /daily, /hourly, /weekly, /monthly(/), /yearly, /top, /oneline")
	
	// Start Grafana Cloud push if configured (after server info, before server starts)
	if *grafanaURL != "" && *grafanaUser != "" && *grafanaToken != "" {
		go startGrafanaPush(*port, *token, *grafanaURL, *grafanaUser, *grafanaToken, *grafanaInterval, service)
		log.Printf("Grafana Cloud push: enabled (interval: %v)", *grafanaInterval)
	} else if *grafanaURL != "" || *grafanaUser != "" || *grafanaToken != "" {
		log.Printf("Warning: Grafana Cloud push partially configured, disabled. All of -grafana-url, -grafana-user, and -grafana-token must be set.")
	}
	
	log.Printf("Press Ctrl+C to stop")

	// Start HTTP server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}

// startGrafanaPush starts a background goroutine to periodically push metrics to Grafana Cloud
func startGrafanaPush(port, token, grafanaURL, grafanaUser, grafanaToken string, interval time.Duration, service *VnstatService) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Wait for HTTP server to be ready before first push
	time.Sleep(2 * time.Second)
	
	// Retry logic for initial connection
	maxRetries := 5
	healthURL := fmt.Sprintf("http://localhost:%s/health", port)
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			break
		}
		if i < maxRetries-1 {
			time.Sleep(1 * time.Second)
		} else {
			log.Printf("Grafana push: HTTP server not ready after %d retries, will retry on next interval", maxRetries)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Track if first push succeeded (for initial success log)
	firstPush := true

	// Push immediately after server is ready
	pushMetrics(client, grafanaURL, grafanaUser, grafanaToken, service, &firstPush)

	// Push periodically
	for range ticker.C {
		pushMetrics(client, grafanaURL, grafanaUser, grafanaToken, service, &firstPush)
	}
}

// pushMetrics fetches metrics and pushes them to Grafana Cloud in Protobuf format
// firstPush is used to log the first successful push, then silence subsequent success logs
func pushMetrics(client *http.Client, grafanaURL, grafanaUser, grafanaToken string, service *VnstatService, firstPush *bool) {
	// Get JSON data directly from service
	jsonData, err := service.GetJSON()
	if err != nil {
		log.Printf("Grafana push: failed to get JSON data: %v", err)
		return
	}

	// Parse JSON and convert to Protobuf
	var vnstatData map[string]interface{}
	if err := json.Unmarshal(jsonData, &vnstatData); err != nil {
		log.Printf("Grafana push: failed to parse JSON data: %v", err)
		return
	}

	// Get hostname for labeling
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		log.Printf("Grafana push: failed to get hostname, using 'unknown': %v", err)
	}

	// Convert to Prometheus Remote Write Protobuf format
	writeRequest := convertToPrometheusWriteRequest(vnstatData, hostname)
	if writeRequest == nil {
		log.Printf("Grafana push: failed to convert metrics to Protobuf format")
		return
	}

	// Marshal to Protobuf (prompb uses gogo/protobuf, has its own Marshal method)
	protoData, err := writeRequest.Marshal()
	if err != nil {
		log.Printf("Grafana push: failed to marshal Protobuf: %v", err)
		return
	}

	// Compress with Snappy
	compressed := snappy.Encode(nil, protoData)

	// Push to Grafana Cloud
	req, err := http.NewRequest("POST", grafanaURL, bytes.NewReader(compressed))
	if err != nil {
		log.Printf("Grafana push: failed to create request: %v", err)
		return
	}
	
	// Set Basic Auth (Instance ID as username, API Token as password)
	req.SetBasicAuth(grafanaUser, grafanaToken)
	
	// Set headers for Prometheus remote write
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	pushResp, err := client.Do(req)
	if err != nil {
		log.Printf("Grafana push: failed to push metrics: %v", err)
		return
	}
	defer pushResp.Body.Close()

	// Log first successful push, then only log failures to avoid log spam
	if pushResp.StatusCode == http.StatusNoContent || pushResp.StatusCode == http.StatusOK {
		if *firstPush {
			log.Printf("Grafana push: metrics pushed successfully (subsequent successful pushes will be silent)")
			*firstPush = false
		}
	} else {
		body, _ := io.ReadAll(pushResp.Body)
		log.Printf("Grafana push: failed (status: %d, response: %s)", pushResp.StatusCode, string(body))
	}
}

// convertToPrometheusWriteRequest converts vnstat JSON data to Prometheus Remote Write Protobuf format
func convertToPrometheusWriteRequest(data map[string]interface{}, hostname string) *prompb.WriteRequest {
	now := time.Now().UnixMilli()
	var timeseries []*prompb.TimeSeries

	interfaces, ok := data["interfaces"].([]interface{})
	if !ok {
		return nil
	}

	for _, iface := range interfaces {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}

		interfaceName := fmt.Sprintf("%v", ifaceMap["name"])
		traffic, ok := ifaceMap["traffic"].(map[string]interface{})
		if !ok {
			continue
		}

		// Total traffic
		if total, ok := traffic["total"].(map[string]interface{}); ok {
			if rx, ok := total["rx"].(float64); ok {
				timeseries = append(timeseries, createTimeSeries(
					"vnstat_traffic_total_bytes",
					map[string]string{"hostname": hostname, "interface": interfaceName, "direction": "rx"},
					rx,
					now,
				))
			}
			if tx, ok := total["tx"].(float64); ok {
				timeseries = append(timeseries, createTimeSeries(
					"vnstat_traffic_total_bytes",
					map[string]string{"hostname": hostname, "interface": interfaceName, "direction": "tx"},
					tx,
					now,
				))
			}
		}

		// Monthly traffic
		if month, ok := traffic["month"].([]interface{}); ok && len(month) > 0 {
			if monthData, ok := month[0].(map[string]interface{}); ok {
				if rx, ok := monthData["rx"].(float64); ok {
					timeseries = append(timeseries, createTimeSeries(
						"vnstat_traffic_month_bytes",
						map[string]string{"hostname": hostname, "interface": interfaceName, "direction": "rx"},
						rx,
						now,
					))
				}
				if tx, ok := monthData["tx"].(float64); ok {
					timeseries = append(timeseries, createTimeSeries(
						"vnstat_traffic_month_bytes",
						map[string]string{"hostname": hostname, "interface": interfaceName, "direction": "tx"},
						tx,
						now,
					))
				}
			}
		}

		// Today's traffic (from day array, last element is today)
		if day, ok := traffic["day"].([]interface{}); ok && len(day) > 0 {
			dayData, ok := day[len(day)-1].(map[string]interface{})
			if ok {
				if rx, ok := dayData["rx"].(float64); ok {
					timeseries = append(timeseries, createTimeSeries(
						"vnstat_traffic_today_bytes",
						map[string]string{"hostname": hostname, "interface": interfaceName, "direction": "rx"},
						rx,
						now,
					))
				}
				if tx, ok := dayData["tx"].(float64); ok {
					timeseries = append(timeseries, createTimeSeries(
						"vnstat_traffic_today_bytes",
						map[string]string{"hostname": hostname, "interface": interfaceName, "direction": "tx"},
						tx,
						now,
					))
				}
			}
		}
	}

	return &prompb.WriteRequest{
		Timeseries: convertTimeSeriesSlice(timeseries),
	}
}

// createTimeSeries creates a Prometheus TimeSeries from metric name, labels, and value
func createTimeSeries(metricName string, labels map[string]string, value float64, timestamp int64) *prompb.TimeSeries {
	// Build labels (prompb.Label is a value type, not pointer)
	promLabels := make([]prompb.Label, 0, len(labels)+1)
	
	// Add __name__ label (metric name)
	promLabels = append(promLabels, prompb.Label{
		Name:  "__name__",
		Value: metricName,
	})
	
	// Add other labels
	for k, v := range labels {
		promLabels = append(promLabels, prompb.Label{
			Name:  k,
			Value: v,
		})
	}

	// Create sample (prompb.Sample is a value type, not pointer)
	sample := prompb.Sample{
		Value:     value,
		Timestamp: timestamp,
	}

	return &prompb.TimeSeries{
		Labels:  promLabels,
		Samples: []prompb.Sample{sample},
	}
}

// convertTimeSeriesSlice converts []*prompb.TimeSeries to []prompb.TimeSeries
func convertTimeSeriesSlice(tsSlice []*prompb.TimeSeries) []prompb.TimeSeries {
	result := make([]prompb.TimeSeries, len(tsSlice))
	for i, ts := range tsSlice {
		result[i] = *ts
	}
	return result
}

