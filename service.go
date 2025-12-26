package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// VnstatService wraps vnstat command execution
type VnstatService struct {
	interfaceName string // Network interface name to query
}

// NewVnstatService creates a new VnstatService instance
func NewVnstatService(interfaceName string) *VnstatService {
	return &VnstatService{
		interfaceName: interfaceName,
	}
}

// GetJSON executes vnstat --json command and returns JSON data
func (s *VnstatService) GetJSON() ([]byte, error) {
	args := []string{"--json"}
	if s.interfaceName != "" {
		args = append(args, "-i", s.interfaceName)
	}

	cmd := exec.Command("vnstat", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if command is not found
		if _, ok := err.(*exec.Error); ok {
			return nil, fmt.Errorf("vnstat is not installed or not in PATH: %v", err)
		}
		// Command execution failed, return error message
		return nil, fmt.Errorf("vnstat execution failed: %s, error: %v", stderr.String(), err)
	}

	// Validate that the returned data is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal(stdout.Bytes(), &jsonData); err != nil {
		return nil, fmt.Errorf("vnstat returned invalid JSON data: %v", err)
	}

	return stdout.Bytes(), nil
}

// executeCommand is a generic method to execute vnstat commands
func (s *VnstatService) executeCommand(args []string) ([]byte, error) {
	if s.interfaceName != "" {
		args = append(args, "-i", s.interfaceName)
	}

	cmd := exec.Command("vnstat", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if command is not found
		if _, ok := err.(*exec.Error); ok {
			return nil, fmt.Errorf("vnstat is not installed or not in PATH: %v", err)
		}
		// Command execution failed, return error message
		return nil, fmt.Errorf("vnstat execution failed: %s, error: %v", stderr.String(), err)
	}

	return stdout.Bytes(), nil
}

// GetText executes vnstat -m command and returns text data (monthly view)
func (s *VnstatService) GetText() ([]byte, error) {
	return s.executeCommand([]string{"-m"})
}

// GetSummary executes vnstat command and returns default summary view
func (s *VnstatService) GetSummary() ([]byte, error) {
	return s.executeCommand([]string{})
}

// GetDaily executes vnstat -d command and returns daily view
func (s *VnstatService) GetDaily() ([]byte, error) {
	return s.executeCommand([]string{"-d"})
}

// GetHourly executes vnstat -h command and returns hourly view
func (s *VnstatService) GetHourly() ([]byte, error) {
	return s.executeCommand([]string{"-h"})
}

// GetWeekly executes vnstat -w command and returns weekly view
func (s *VnstatService) GetWeekly() ([]byte, error) {
	return s.executeCommand([]string{"-w"})
}

// GetYearly executes vnstat -y command and returns yearly view
func (s *VnstatService) GetYearly() ([]byte, error) {
	return s.executeCommand([]string{"-y"})
}

// GetTop executes vnstat -t command and returns top traffic interfaces
func (s *VnstatService) GetTop() ([]byte, error) {
	return s.executeCommand([]string{"-t"})
}

// GetOneline executes vnstat --oneline command and returns one-line output
func (s *VnstatService) GetOneline() ([]byte, error) {
	return s.executeCommand([]string{"--oneline"})
}

// CheckVnstatInstalled checks if vnstat is installed
func (s *VnstatService) CheckVnstatInstalled() error {
	cmd := exec.Command("vnstat", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vnstat is not installed or not in PATH")
	}
	return nil
}

