package garmin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL       = "https://connect.garmin.com"
	ssoURL        = "https://sso.garmin.com/sso"
	signinURL     = ssoURL + "/signin"
	activityURL   = baseURL + "/modern/proxy/activitylist-service/activities/search/activities"
)

// Activity represents a Garmin activity
type Activity struct {
	ActivityID   int64     `json:"activityId"`
	ActivityName string    `json:"activityName"`
	ActivityType string    `json:"activityType"`
	StartTime    time.Time `json:"startTimeLocal"`
	Distance     float64   `json:"distance"`      // meters
	Duration     float64   `json:"duration"`      // seconds
	AvgSpeed     float64   `json:"averageSpeed"`  // m/s
	Calories     float64   `json:"calories"`
}

// Client is a Garmin Connect API client
type Client struct {
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new Garmin Connect client
func NewClient(username, password string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		username: username,
		password: password,
		httpClient: &http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
		},
	}
}

// Login authenticates with Garmin Connect
func (c *Client) Login() error {
	// Step 1: Get CSRF token
	req, err := http.NewRequest("GET", signinURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get login page: %w", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Step 2: Perform login
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)
	data.Set("embed", "false")

	req, err = http.NewRequest("POST", signinURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	// Check if login was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Step 3: Follow redirects and complete authentication
	req, err = http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return err
	}

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to complete authentication: %w", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return nil
}

// GetActivities retrieves activities for a specific date
func (c *Client) GetActivities(date time.Time) ([]Activity, error) {
	startDate := date.Format("2006-01-02")
	endDate := date.Add(24 * time.Hour).Format("2006-01-02")

	// Build request
	params := map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
		"limit":     100,
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", activityURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("NK", "NT") // Required header for Garmin API

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get activities: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The API returns an object with activities nested inside
	var response struct {
		Activities []Activity `json:"activityList"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		// Debug: print raw response if decode fails
		fmt.Printf("Raw response: %s\n", string(body))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Activities, nil
}

// GetBikeActivities filters and returns only cycling activities
func (c *Client) GetBikeActivities(date time.Time) ([]Activity, error) {
	activities, err := c.GetActivities(date)
	if err != nil {
		return nil, err
	}

	var bikeActivities []Activity
	for _, activity := range activities {
		activityType := strings.ToLower(activity.ActivityType)
		if strings.Contains(activityType, "cycling") ||
		   strings.Contains(activityType, "bike") ||
		   strings.Contains(activityType, "biking") {
			bikeActivities = append(bikeActivities, activity)
		}
	}

	return bikeActivities, nil
}

// DownloadActivity downloads the full activity data (GPX, TCX, or FIT format)
func (c *Client) DownloadActivity(activityID int64) ([]byte, error) {
	downloadURL := fmt.Sprintf("%s/download-service/export/gpx/activity/%d", baseURL, activityID)

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download activity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download activity: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read activity data: %w", err)
	}

	return data, nil
}

// Logout logs out from Garmin Connect
func (c *Client) Logout() error {
	// Garmin logout is not strictly necessary as we're using session cookies
	// but it's good practice to clear the session
	return nil
}
