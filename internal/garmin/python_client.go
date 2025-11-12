package garmin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// PythonClient wraps the Python garminconnect library
type PythonClient struct {
	username   string
	password   string
	scriptPath string
}

// NewPythonClient creates a new Python-based Garmin client
func NewPythonClient(username, password string) *PythonClient {
	return &PythonClient{
		username: username,
		password: password,
	}
}

// Login is a no-op for Python client (login happens per request)
func (c *PythonClient) Login() error {
	// Extract embedded script to temp file if not already done
	if c.scriptPath == "" {
		tmpFile, err := os.CreateTemp("", "garmin_client_*.py")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer tmpFile.Close()

		if _, err := tmpFile.Write(pythonScript); err != nil {
			return fmt.Errorf("failed to write Python script: %w", err)
		}

		c.scriptPath = tmpFile.Name()
	}

	// Test that python3 is available
	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 not found in PATH")
	}

	return nil
}

// GetActivities retrieves activities for a specific date using Python script
func (c *PythonClient) GetActivities(date time.Time) ([]Activity, error) {
	dateStr := date.Format("2006-01-02")

	// Get absolute path to script
	absScriptPath, err := filepath.Abs(c.scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute script path: %w", err)
	}

	cmd := exec.Command("python3", absScriptPath,
		"--username", c.username,
		"--password", c.password,
		"--command", "get-activities",
		"--date", dateStr)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("Python script failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run Python script: %w", err)
	}

	var activities []Activity
	if err := json.Unmarshal(output, &activities); err != nil {
		return nil, fmt.Errorf("failed to parse activities JSON: %w", err)
	}

	return activities, nil
}

// GetBikeActivities returns bike activities (already filtered by Python script)
func (c *PythonClient) GetBikeActivities(date time.Time) ([]Activity, error) {
	return c.GetActivities(date)
}

// DownloadActivity downloads activity in FIT format using Python script
func (c *PythonClient) DownloadActivity(activityID int64) ([]byte, error) {
	// Get absolute path to script
	absScriptPath, err := filepath.Abs(c.scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute script path: %w", err)
	}

	cmd := exec.Command("python3", absScriptPath,
		"--username", c.username,
		"--password", c.password,
		"--command", "download-activity",
		"--activity-id", fmt.Sprintf("%d", activityID),
		"--format", "FIT")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("Python script failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run Python script: %w", err)
	}

	return output, nil
}

// Logout cleans up the temporary Python script file
func (c *PythonClient) Logout() error {
	if c.scriptPath != "" {
		// Best effort cleanup - ignore errors
		os.Remove(c.scriptPath)
		c.scriptPath = ""
	}
	return nil
}
