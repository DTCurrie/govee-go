package govee_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	govee "github.com/DTCurrie/govee-go"
)

func TestNew(t *testing.T) {
	t.Parallel()

	c := govee.New("my-key")
	if c == nil {
		t.Fatal("New returned nil")
	}
}

func TestNew_WithOptions(t *testing.T) {
	t.Parallel()

	custom := &http.Client{}
	c := govee.New("my-key",
		govee.WithBaseURL("http://localhost:9999"),
		govee.WithHTTPClient(custom),
	)
	if c == nil {
		t.Fatal("New returned nil")
	}
}

func TestGetDevices(t *testing.T) {
	t.Parallel()

	want := []govee.Device{
		{
			SKU:        "H6008",
			DeviceID:   "AA:BB:CC:DD",
			DeviceName: "Bedroom Light",
			Type:       govee.DeviceLight,
			Capabilities: []govee.Capability{
				{Type: govee.CapabilityOnOff, Instance: "powerSwitch"},
				{Type: govee.CapabilityRange, Instance: "brightness"},
			},
		},
		{
			SKU:      "H5083",
			DeviceID: "11:22:33:44",
			Type:     govee.DeviceSocket,
			Capabilities: []govee.Capability{
				{Type: govee.CapabilityOnOff, Instance: "powerSwitch"},
			},
		},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/user/devices" {
			t.Errorf("expected path /user/devices, got %s", r.URL.Path)
		}
		assertHeader(t, r, "Govee-API-Key", "test-api-key")
		respondGetJSON(t, w, want)
	}))

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		t.Fatalf("GetDevices: %v", err)
	}
	if len(devices) != len(want) {
		t.Fatalf("device count: got %d, want %d", len(devices), len(want))
	}
	if devices[0].SKU != want[0].SKU {
		t.Errorf("devices[0].SKU: got %q, want %q", devices[0].SKU, want[0].SKU)
	}
	if devices[0].DeviceID != want[0].DeviceID {
		t.Errorf("devices[0].DeviceID: got %q, want %q", devices[0].DeviceID, want[0].DeviceID)
	}
	if len(devices[0].Capabilities) != len(want[0].Capabilities) {
		t.Errorf("devices[0].Capabilities count: got %d, want %d",
			len(devices[0].Capabilities), len(want[0].Capabilities))
	}
}

func TestGetDevices_APIError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondAPIError(t, w, http.StatusTooManyRequests, "Rate limit exceeded")
	}))

	_, err := client.GetDevices(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*govee.APIError)
	if !ok {
		t.Fatalf("expected *govee.APIError, got %T", err)
	}
	if apiErr.Code != http.StatusTooManyRequests {
		t.Errorf("Code: got %d, want %d", apiErr.Code, http.StatusTooManyRequests)
	}
	if apiErr.Message != "Rate limit exceeded" {
		t.Errorf("Message: got %q, want %q", apiErr.Message, "Rate limit exceeded")
	}
}

func TestGetDevices_EnvelopeError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type envelope struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(envelope{
			Code:    http.StatusUnauthorized,
			Message: "Invalid API key",
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))

	_, err := client.GetDevices(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*govee.APIError)
	if !ok {
		t.Fatalf("expected *govee.APIError, got %T", err)
	}
	if apiErr.Code != http.StatusUnauthorized {
		t.Errorf("Code: got %d, want %d", apiErr.Code, http.StatusUnauthorized)
	}
}

func TestGetDevices_Cancelled(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondGetJSON(t, w, []govee.Device{})
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetDevices(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestDevice_FindCapability(t *testing.T) {
	t.Parallel()

	d := govee.Device{
		SKU:      "H6008",
		DeviceID: "AA:BB",
		Capabilities: []govee.Capability{
			{Type: govee.CapabilityOnOff, Instance: "powerSwitch"},
			{Type: govee.CapabilityRange, Instance: "brightness"},
			{Type: govee.CapabilityColorSetting, Instance: "colorRgb"},
		},
	}

	tests := []struct {
		name     string
		capType  string
		instance string
		wantNil  bool
	}{
		{"found on_off", govee.CapabilityOnOff, "powerSwitch", false},
		{"found brightness", govee.CapabilityRange, "brightness", false},
		{"not found unknown type", "unknown.type", "instance", true},
		{"not found wrong instance", govee.CapabilityOnOff, "wrongInstance", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := d.FindCapability(tc.capType, tc.instance)
			if tc.wantNil && got != nil {
				t.Errorf("expected nil, got %v", got)
			}
			if !tc.wantNil && got == nil {
				t.Errorf("expected non-nil capability")
			}
		})
	}
}

func TestDevice_HasCapability(t *testing.T) {
	t.Parallel()

	d := govee.Device{
		Capabilities: []govee.Capability{
			{Type: govee.CapabilityOnOff, Instance: "powerSwitch"},
		},
	}

	if !d.HasCapability(govee.CapabilityOnOff, "powerSwitch") {
		t.Error("expected HasCapability to return true for powerSwitch")
	}
	if d.HasCapability(govee.CapabilityRange, "brightness") {
		t.Error("expected HasCapability to return false for brightness")
	}
}

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	err := &govee.APIError{Code: 401, Message: "Unauthorized"}
	want := "govee: API error 401: Unauthorized"
	if got := err.Error(); got != want {
		t.Errorf("Error(): got %q, want %q", got, want)
	}
}
