package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// TokenConfig represents the configuration for token retrieval
type TokenConfig struct {
	BaseURL        string
	ClientID       string
	ServiceAccount string
	ServiceToken   string
	SkipTLSVerify  bool
}

// TokenResponse represents the structure of the token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewTokenConfig creates a configuration from environment variables
func NewTokenConfig() *TokenConfig {
	return &TokenConfig{
		BaseURL:        getEnv("BASE_URL", "https://localhost:9443"),
		ClientID:       getEnv("CLIENT_ID", ""),
		ServiceAccount: getEnv("SERVICE_ACCOUNT", ""),
		ServiceToken:   getEnv("SERVICE_TOKEN", ""),
		SkipTLSVerify:  getBoolEnv("SKIP_TLS_VERIFY", false),
	}
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getBoolEnv retrieves a boolean environment variable
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	switch strings.ToLower(value) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

// getAccessToken retrieves an access token from the specified endpoint
func getAccessToken(config *TokenConfig) (*TokenResponse, error) {
	// Validate required configuration
	if config.ClientID == "" || config.ServiceAccount == "" || config.ServiceToken == "" {
		return nil, fmt.Errorf("missing required configuration")
	}

	// Create a custom HTTP client with optional TLS verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipTLSVerify},
	}
	client := &http.Client{Transport: tr}

	// Prepare the request body
	data := url.Values{
		"grant_type": {"client_credentials"},
		"client_id":  {config.ClientID},
		"username":   {config.ServiceAccount},
		"password":   {config.ServiceToken},
		"scope":      {"profile"},
	}

	// Create the HTTP request
	fullURL := strings.TrimRight(config.BaseURL, "/") + "/application/o/token/"
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token retrieval failed (status %d): %s",
			resp.StatusCode, string(body))
	}

	// Parse the JSON response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("error parsing token response: %w\nRaw response: %s", err, string(body))
	}

	return &tokenResp, nil
}

func main() {
	// Load configuration from environment
	config := NewTokenConfig()

	// Retrieve access token
	tokenResp, err := getAccessToken(config)
	if err != nil {
		log.Fatalf("Failed to obtain access token: %v", err)
	}

	// Print token details
	fmt.Println("Access Token:", tokenResp.AccessToken)
	fmt.Println("Token Type:", tokenResp.TokenType)
	fmt.Println("Expires In:", tokenResp.ExpiresIn, "seconds")
}
