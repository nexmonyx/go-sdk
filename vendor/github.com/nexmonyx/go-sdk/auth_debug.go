package nexmonyx

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DebugAuthHeaders performs test requests to validate authentication headers
// This helps diagnose authentication issues by testing both header formats
func (c *Client) DebugAuthHeaders(ctx context.Context) error {
	fmt.Println("=== Nexmonyx SDK Authentication Debug ===")
	fmt.Printf("SDK Version: %s\n", Version)
	fmt.Printf("Base URL: %s\n", c.config.BaseURL)
	fmt.Println()

	// Check current authentication configuration
	fmt.Println("Current Authentication Configuration:")
	if c.config.Auth.Token != "" {
		fmt.Println("  Auth Type: JWT Token")
	} else if c.config.Auth.APIKey != "" && c.config.Auth.APISecret != "" {
		fmt.Println("  Auth Type: API Key/Secret")
		fmt.Printf("  API Key: %s\n", c.config.Auth.APIKey)
	} else if c.config.Auth.ServerUUID != "" && c.config.Auth.ServerSecret != "" {
		fmt.Println("  Auth Type: Server Credentials")
		fmt.Printf("  Server UUID: %s\n", c.config.Auth.ServerUUID)
	} else if c.config.Auth.MonitoringKey != "" {
		fmt.Println("  Auth Type: Monitoring Key")
	} else {
		fmt.Println("  Auth Type: None (no authentication configured)")
	}
	fmt.Println()

	// Only proceed with header tests if we have server credentials
	if c.config.Auth.ServerUUID == "" || c.config.Auth.ServerSecret == "" {
		fmt.Println("Skipping header format tests - no server credentials configured")
		return nil
	}

	fmt.Println("Testing Authentication Header Formats:")
	fmt.Println("----------------------------------------")

	// Test 1: Current SDK format (X-Server-UUID, X-Server-Secret)
	fmt.Println("\nTest 1: Headers with X- prefix (SDK default)")
	err1 := c.testAuthRequest(ctx, map[string]string{
		"X-Server-UUID":   c.config.Auth.ServerUUID,
		"X-Server-Secret": c.config.Auth.ServerSecret,
	})
	if err1 != nil {
		fmt.Printf("  Result: FAILED - %v\n", err1)
	} else {
		fmt.Println("  Result: SUCCESS")
	}

	// Test 2: Without X- prefix (Server-UUID, Server-Secret)
	fmt.Println("\nTest 2: Headers without X- prefix")
	err2 := c.testAuthRequest(ctx, map[string]string{
		"Server-UUID":   c.config.Auth.ServerUUID,
		"Server-Secret": c.config.Auth.ServerSecret,
	})
	if err2 != nil {
		fmt.Printf("  Result: FAILED - %v\n", err2)
	} else {
		fmt.Println("  Result: SUCCESS")
	}

	fmt.Println("\n========================================")

	// Summary
	if err1 == nil && err2 != nil {
		fmt.Println("CONCLUSION: API expects headers WITH 'X-' prefix (current SDK format is correct)")
		return nil
	} else if err1 != nil && err2 == nil {
		fmt.Println("CONCLUSION: API expects headers WITHOUT 'X-' prefix")
		fmt.Println("ACTION REQUIRED: SDK needs to be updated to use 'Server-UUID' instead of 'X-Server-UUID'")
		return fmt.Errorf("header format mismatch detected")
	} else if err1 == nil && err2 == nil {
		fmt.Println("CONCLUSION: API accepts both header formats")
		return nil
	} else {
		fmt.Println("CONCLUSION: Neither header format worked - authentication issue may be elsewhere")
		return fmt.Errorf("authentication failed with both header formats")
	}
}

// testAuthRequest performs a test request with custom headers
func (c *Client) testAuthRequest(ctx context.Context, headers map[string]string) error {
	// Create a custom HTTP client for this test
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/v1/heartbeat", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set standard headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	// Set custom auth headers
	for k, v := range headers {
		req.Header.Set(k, v)
		if k == "X-Server-Secret" || k == "Server-Secret" {
			fmt.Printf("  Setting %s: [REDACTED]\n", k)
		} else {
			fmt.Printf("  Setting %s: %s\n", k, v)
		}
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("  Response Status: %d %s\n", resp.StatusCode, resp.Status)

	// Read response body for error details
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		if n > 0 {
			fmt.Printf("  Error Details: %s\n", string(body[:n]))
		}
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return nil
}
