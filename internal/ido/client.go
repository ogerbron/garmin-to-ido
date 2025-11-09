package ido

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	idoBaseURL  = "https://www.idosport.app"
	loginURL    = idoBaseURL + "/login"
	athleteURL  = idoBaseURL + "/athlete/"
	uploadURL   = idoBaseURL + "/athlete/activity/add"
)

// Client is an iDO Sport client using browser automation
type Client struct {
	username string
	password string
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewClient creates a new iDO Sport client
func NewClient(username, password string) (*Client, error) {
	// Create chrome context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)

	return &Client{
		username: username,
		password: password,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Login logs into iDO Sport
func (c *Client) Login() error {
	var pageURL string

	err := chromedp.Run(c.ctx,
		// Navigate to login page
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(`input[name="email"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),

		// Fill in the form (this will ensure CSRF token is part of the form)
		chromedp.SendKeys(`input[name="email"]`, c.username, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="password"]`, c.password, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		// Submit the form using the form's submit method
		chromedp.Evaluate(`document.querySelector('form').submit()`, nil),
		chromedp.Sleep(6*time.Second), // Wait for login to complete and redirect

		// Check current URL
		chromedp.Location(&pageURL),
	)

	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	// Check if we're redirected away from login page (successful login)
	// Could be athlete page, structure page, or other authenticated pages
	if pageURL == loginURL || pageURL == loginURL+"/" {
		return fmt.Errorf("login failed - still on login page (check credentials)")
	}

	// If we're not on login page anymore, navigate to athlete page
	if pageURL != athleteURL && pageURL != athleteURL+"/" {
		err = chromedp.Run(c.ctx,
			chromedp.Navigate(athleteURL),
			chromedp.Sleep(2*time.Second),
		)
		if err != nil {
			return fmt.Errorf("failed to navigate to athlete page: %w", err)
		}
	}

	return nil
}

// UploadActivity uploads an activity to iDO Sport using the API
func (c *Client) UploadActivity(activityData []byte, activityName string) error {
	fmt.Printf("Uploading activity: %s (%d bytes)\n", activityName, len(activityData))

	// Get cookies from browser session for idosport.app domain
	var cookies []*network.Cookie
	if err := chromedp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		// Get all cookies
		allCookies, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		// Filter cookies for idosport.app
		for _, cookie := range allCookies {
			if cookie.Domain == ".idosport.app" || cookie.Domain == "www.idosport.app" || cookie.Domain == "idosport.app" {
				cookies = append(cookies, cookie)
			}
		}
		return nil
	})); err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}

	// Build cookie string and check for PHPSESSID
	var cookieStr string
	hasSession := false
	for i, cookie := range cookies {
		if i > 0 {
			cookieStr += "; "
		}
		cookieStr += cookie.Name + "=" + cookie.Value
		if cookie.Name == "PHPSESSID" {
			hasSession = true
		}
	}

	if !hasSession {
		return fmt.Errorf("no session cookie found - login may have failed")
	}

	// Step 1: Get S3 upload URL
	req, err := http.NewRequest("GET", idoBaseURL+"/v-get-s3-s-upurl", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6")
	req.Header.Set("Referer", athleteURL)
	req.Header.Set("Origin", idoBaseURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get upload URL: %d %s", resp.StatusCode, string(body))
	}

	var s3Response struct {
		URL string `json:"url"`
		Key string `json:"key"`
	}

	if err := json.Unmarshal(body, &s3Response); err != nil {
		return fmt.Errorf("failed to parse S3 response: %w", err)
	}

	// Step 2: Upload file to S3
	s3Req, err := http.NewRequest("PUT", s3Response.URL, bytes.NewReader(activityData))
	if err != nil {
		return fmt.Errorf("failed to create S3 request: %w", err)
	}

	s3Req.Header.Set("Content-Type", "application/octet-stream")

	s3Resp, err := client.Do(s3Req)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	defer s3Resp.Body.Close()

	if s3Resp.StatusCode != 200 {
		body, _ := io.ReadAll(s3Resp.Body)
		return fmt.Errorf("S3 upload failed: %d %s", s3Resp.StatusCode, string(body))
	}

	// Step 3: Create activity record
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("actName", activityName)
	writer.WriteField("dateString", time.Now().Format("2006-01-02"))
	writer.WriteField("creationType", "fit")
	writer.WriteField("sportType", "bike")
	writer.WriteField("keyFile", s3Response.Key)

	writer.Close()

	activityReq, err := http.NewRequest("POST", idoBaseURL+"/v-add-activity-v2", &buf)
	if err != nil {
		return fmt.Errorf("failed to create activity request: %w", err)
	}

	activityReq.Header.Set("Cookie", cookieStr)
	activityReq.Header.Set("Content-Type", writer.FormDataContentType())
	activityReq.Header.Set("Accept", "application/json, text/plain, */*")
	activityReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	activityReq.Header.Set("Origin", idoBaseURL)
	activityReq.Header.Set("Referer", athleteURL)

	activityResp, err := client.Do(activityReq)
	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	defer activityResp.Body.Close()

	if activityResp.StatusCode != 200 {
		body, _ := io.ReadAll(activityResp.Body)
		return fmt.Errorf("activity creation failed: %d %s", activityResp.StatusCode, string(body))
	}

	return nil
}

// Close closes the browser and cleans up resources
func (c *Client) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
