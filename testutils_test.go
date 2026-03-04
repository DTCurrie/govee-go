package govee_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	govee "github.com/DTCurrie/govee-go"
)

// newTestClient creates a Client pointed at an httptest.Server backed by handler.
// The server is closed via t.Cleanup.
func newTestClient(t *testing.T, handler http.Handler) *govee.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return govee.New("test-api-key", govee.WithBaseURL(srv.URL))
}

// respondGetJSON writes a GET-style response envelope: {code:200, message:"success", data:v}.
func respondGetJSON(t *testing.T, w http.ResponseWriter, v interface{}) {
	t.Helper()
	type envelope struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(envelope{Code: 200, Message: "success", Data: v}); err != nil {
		t.Errorf("respondGetJSON encode: %v", err)
	}
}

// respondPostJSON writes a POST-style response envelope: {code:200, msg:"success", payload:v}.
func respondPostJSON(t *testing.T, w http.ResponseWriter, v interface{}) {
	t.Helper()
	type envelope struct {
		RequestID string      `json:"requestId"`
		Code      int         `json:"code"`
		Msg       string      `json:"msg"`
		Payload   interface{} `json:"payload"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(envelope{Code: 200, Msg: "success", Payload: v}); err != nil {
		t.Errorf("respondPostJSON encode: %v", err)
	}
}

// respondAPIError writes a GET-style error envelope with the given HTTP status.
func respondAPIError(t *testing.T, w http.ResponseWriter, statusCode int, msg string) {
	t.Helper()
	type envelope struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(envelope{Code: statusCode, Message: msg}); err != nil {
		t.Errorf("respondAPIError encode: %v", err)
	}
}

// decodePostBody reads and decodes the JSON body of an incoming POST request.
func decodePostBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("decodePostBody: %v", err)
	}
	return body
}

// postPayload extracts the "payload" field from a decoded request body.
func postPayload(t *testing.T, body map[string]interface{}) map[string]interface{} {
	t.Helper()
	p, ok := body["payload"].(map[string]interface{})
	if !ok {
		t.Errorf("payload field not found or not an object; body = %v", body)
	}
	return p
}

// assertHeader fails the test if the request header key does not equal want.
func assertHeader(t *testing.T, r *http.Request, key, want string) {
	t.Helper()
	if got := r.Header.Get(key); got != want {
		t.Errorf("header %q: got %q, want %q", key, got, want)
	}
}
