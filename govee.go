// Package govee provides a Go client for the Govee OpenAPI.
//
// The Govee OpenAPI uses a capability-based model: each device reports the
// capabilities it supports (on/off, brightness, color, scenes, etc.) along with
// valid parameter ranges and options. All control commands go through a single
// endpoint using a {type, instance, value} triple.
//
// # Getting started
//
//	client := govee.New("your-api-key")
//	devices, err := client.GetDevices(ctx)
//	err = client.TurnOn(ctx, devices[0].SKU, devices[0].DeviceID)
//
// # Scenes
//
// Static scenes are included in the device capabilities from GetDevices.
// Dynamic scenes (which can be numerous) must be fetched separately:
//
//	scenes, err := client.GetScenes(ctx, sku, deviceID)
//	err = client.SetLightScene(ctx, sku, deviceID, scenes[0].Value)
//
// # Events
//
// To receive real-time device events (sensor readings, alerts), use EventClient:
//
//	ec := govee.NewEventClient("your-api-key", func(event govee.DeviceEvent) {
//	    fmt.Printf("event from %s: %v\n", event.SKU, event.Capabilities)
//	})
//	err := ec.Connect(ctx)
//	defer ec.Close()
package govee

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://openapi.api.govee.com/router/api/v1"

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

// getEnvelope is the response envelope for GET endpoints.
type getEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// postEnvelope is the response envelope for POST endpoints.
type postEnvelope struct {
	RequestID string          `json:"requestId"`
	Code      int             `json:"code"`
	Msg       string          `json:"msg"`
	Message   string          `json:"message"`
	Payload   json.RawMessage `json:"payload"`
}

// apiRequest is the request envelope for POST endpoints.
type apiRequest struct {
	RequestID string      `json:"requestId"`
	Payload   interface{} `json:"payload"`
}

// newRequestID generates a random UUID v4 string for use as a request ID.
func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// doGet performs an authenticated GET request and returns the decoded data field.
func (c *Client) doGet(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Govee-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("govee: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("govee: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var env getEnvelope
		if jsonErr := json.Unmarshal(body, &env); jsonErr == nil && env.Message != "" {
			return nil, &APIError{Code: resp.StatusCode, Message: env.Message}
		}
		return nil, &APIError{Code: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	var env getEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("govee: failed to decode response: %w", err)
	}
	if env.Code != http.StatusOK {
		return nil, &APIError{Code: env.Code, Message: env.Message}
	}
	return env.Data, nil
}

// doPost performs an authenticated POST request with the given payload and returns
// the decoded payload field from the response envelope.
func (c *Client) doPost(ctx context.Context, path string, payload interface{}) (json.RawMessage, error) {
	reqBody, err := json.Marshal(apiRequest{
		RequestID: newRequestID(),
		Payload:   payload,
	})
	if err != nil {
		return nil, fmt.Errorf("govee: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Govee-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("govee: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("govee: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var env postEnvelope
		if jsonErr := json.Unmarshal(respBody, &env); jsonErr == nil {
			msg := env.Msg
			if msg == "" {
				msg = env.Message
			}
			if msg != "" {
				return nil, &APIError{Code: resp.StatusCode, Message: msg}
			}
		}
		return nil, &APIError{Code: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	var env postEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("govee: failed to decode response: %w", err)
	}
	if env.Code != http.StatusOK {
		msg := env.Msg
		if msg == "" {
			msg = env.Message
		}
		return nil, &APIError{Code: env.Code, Message: msg}
	}
	return env.Payload, nil
}
