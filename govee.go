// Package govee provides a Go client for the Govee Developer REST API v1.
// It supports listing devices, querying device state, and sending control
// commands (on/off, brightness, RGB color, color temperature).
//
// Usage:
//
//	client := govee.New("your-api-key")
//	devices, err := client.GetDevices(ctx)
//	err = client.TurnOn(ctx, devices[0].DeviceID, devices[0].Model)
package govee

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://developer-api.govee.com"

// Client is a Govee API client. Create one with New().
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL. Useful for testing.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient replaces the default *http.Client. Useful for custom transports or timeouts.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// New creates a new Govee API client authenticated with the given API key.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// apiResponse is the common envelope wrapping all Govee API responses.
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doGet performs a GET request and decodes the response envelope.
// It returns the raw data field for further decoding by the caller.
func (c *Client) doGet(ctx context.Context, path string, query map[string]string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Govee-API-Key", c.apiKey)

	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return c.doRequest(req)
}

// doPut performs a PUT request with a JSON body and decodes the response envelope.
func (c *Client) doPut(ctx context.Context, path string, body any) (json.RawMessage, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("govee: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Govee-API-Key", c.apiKey)

	return c.doRequest(req)
}

// doRequest executes an HTTP request, checks the status code, and decodes the
// response envelope, returning the raw data field.
func (c *Client) doRequest(req *http.Request) (json.RawMessage, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("govee: request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("govee: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var envelope apiResponse
		if jsonErr := json.Unmarshal(bodyBytes, &envelope); jsonErr == nil && envelope.Message != "" {
			return nil, &APIError{Code: resp.StatusCode, Message: envelope.Message}
		}
		return nil, &APIError{Code: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	var envelope apiResponse
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
		return nil, fmt.Errorf("govee: failed to decode response: %w", err)
	}

	if envelope.Code != http.StatusOK {
		return nil, &APIError{Code: envelope.Code, Message: envelope.Message}
	}

	return envelope.Data, nil
}
