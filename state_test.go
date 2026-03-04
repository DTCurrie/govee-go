package govee_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	govee "github.com/DTCurrie/govee-go"
)

func TestGetDeviceState(t *testing.T) {
	t.Parallel()

	statePayload := map[string]interface{}{
		"sku":    "H6008",
		"device": "AA:BB:CC:DD",
		"capabilities": []map[string]interface{}{
			{
				"type":     govee.CapabilityOnOff,
				"instance": "powerSwitch",
				"state":    map[string]interface{}{"value": 1},
			},
			{
				"type":     govee.CapabilityRange,
				"instance": "brightness",
				"state":    map[string]interface{}{"value": 75},
			},
		},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/state" {
			t.Errorf("expected /device/state, got %s", r.URL.Path)
		}
		body := decodePostBody(t, r)
		p := postPayload(t, body)
		if p["sku"] != "H6008" {
			t.Errorf("sku: got %v, want H6008", p["sku"])
		}
		if p["device"] != "AA:BB:CC:DD" {
			t.Errorf("device: got %v, want AA:BB:CC:DD", p["device"])
		}
		respondPostJSON(t, w, statePayload)
	}))

	resp, err := client.GetDeviceState(context.Background(), "H6008", "AA:BB:CC:DD")
	if err != nil {
		t.Fatalf("GetDeviceState: %v", err)
	}
	if resp.SKU != "H6008" {
		t.Errorf("SKU: got %q, want H6008", resp.SKU)
	}
	if len(resp.Capabilities) != 2 {
		t.Errorf("capabilities count: got %d, want 2", len(resp.Capabilities))
	}
}

func makeStateResponse(caps ...govee.Capability) *govee.DeviceStateResponse {
	return &govee.DeviceStateResponse{
		SKU:          "H6008",
		DeviceID:     "AA:BB",
		Capabilities: caps,
	}
}

func stateCapability(capType, instance string, value interface{}) govee.Capability {
	raw, _ := json.Marshal(value)
	return govee.Capability{
		Type:     capType,
		Instance: instance,
		State:    &govee.CapabilityState{Value: json.RawMessage(raw)},
	}
}

func TestDeviceStateResponse_FindState(t *testing.T) {
	t.Parallel()

	resp := makeStateResponse(
		stateCapability(govee.CapabilityOnOff, "powerSwitch", 1),
		stateCapability(govee.CapabilityRange, "brightness", 50),
	)

	if got := resp.FindState(govee.CapabilityOnOff, "powerSwitch"); got == nil {
		t.Error("expected to find powerSwitch, got nil")
	}
	if got := resp.FindState("unknown", "instance"); got != nil {
		t.Errorf("expected nil for unknown capability, got %v", got)
	}
}

func TestDeviceStateResponse_IsOnline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		resp  *govee.DeviceStateResponse
		want  bool
	}{
		{
			"online true",
			makeStateResponse(stateCapability(govee.CapabilityOnline, "online", true)),
			true,
		},
		{
			"online false",
			makeStateResponse(stateCapability(govee.CapabilityOnline, "online", false)),
			false,
		},
		{
			"no online capability",
			makeStateResponse(stateCapability(govee.CapabilityOnOff, "powerSwitch", 1)),
			false,
		},
		{
			"online capability without state",
			makeStateResponse(govee.Capability{Type: govee.CapabilityOnline, Instance: "online"}),
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.resp.IsOnline(); got != tc.want {
				t.Errorf("IsOnline(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDeviceStateResponse_PowerState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resp      *govee.DeviceStateResponse
		wantOn    bool
		wantFound bool
	}{
		{
			"power on",
			makeStateResponse(stateCapability(govee.CapabilityOnOff, "powerSwitch", 1)),
			true, true,
		},
		{
			"power off",
			makeStateResponse(stateCapability(govee.CapabilityOnOff, "powerSwitch", 0)),
			false, true,
		},
		{
			"no power capability",
			makeStateResponse(stateCapability(govee.CapabilityRange, "brightness", 50)),
			false, false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			on, found := tc.resp.PowerState()
			if on != tc.wantOn {
				t.Errorf("PowerState() on: got %v, want %v", on, tc.wantOn)
			}
			if found != tc.wantFound {
				t.Errorf("PowerState() found: got %v, want %v", found, tc.wantFound)
			}
		})
	}
}

func TestDeviceStateResponse_Brightness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		resp       *govee.DeviceStateResponse
		wantPct    int
		wantFound  bool
	}{
		{
			"brightness 75",
			makeStateResponse(stateCapability(govee.CapabilityRange, "brightness", 75)),
			75, true,
		},
		{
			"no brightness capability",
			makeStateResponse(stateCapability(govee.CapabilityOnOff, "powerSwitch", 1)),
			0, false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pct, found := tc.resp.Brightness()
			if pct != tc.wantPct {
				t.Errorf("Brightness() pct: got %d, want %d", pct, tc.wantPct)
			}
			if found != tc.wantFound {
				t.Errorf("Brightness() found: got %v, want %v", found, tc.wantFound)
			}
		})
	}
}

func TestDeviceStateResponse_ColorRGB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		resp       *govee.DeviceStateResponse
		wantR, wantG, wantB int
		wantFound  bool
	}{
		{
			"red",
			makeStateResponse(stateCapability(govee.CapabilityColorSetting, "colorRgb", 0xFF0000)),
			255, 0, 0, true,
		},
		{
			"green",
			makeStateResponse(stateCapability(govee.CapabilityColorSetting, "colorRgb", 0x00FF00)),
			0, 255, 0, true,
		},
		{
			"blue",
			makeStateResponse(stateCapability(govee.CapabilityColorSetting, "colorRgb", 0x0000FF)),
			0, 0, 255, true,
		},
		{
			"no color capability",
			makeStateResponse(stateCapability(govee.CapabilityOnOff, "powerSwitch", 1)),
			0, 0, 0, false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r, g, b, found := tc.resp.ColorRGB()
			if r != tc.wantR || g != tc.wantG || b != tc.wantB {
				t.Errorf("ColorRGB(): got (%d,%d,%d), want (%d,%d,%d)",
					r, g, b, tc.wantR, tc.wantG, tc.wantB)
			}
			if found != tc.wantFound {
				t.Errorf("ColorRGB() found: got %v, want %v", found, tc.wantFound)
			}
		})
	}
}

func TestDeviceStateResponse_ColorTemp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resp      *govee.DeviceStateResponse
		wantK     int
		wantFound bool
	}{
		{
			"6500K",
			makeStateResponse(stateCapability(govee.CapabilityColorSetting, "colorTemperatureK", 6500)),
			6500, true,
		},
		{
			"no colorTemp capability",
			makeStateResponse(stateCapability(govee.CapabilityOnOff, "powerSwitch", 1)),
			0, false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			k, found := tc.resp.ColorTemp()
			if k != tc.wantK {
				t.Errorf("ColorTemp() kelvin: got %d, want %d", k, tc.wantK)
			}
			if found != tc.wantFound {
				t.Errorf("ColorTemp() found: got %v, want %v", found, tc.wantFound)
			}
		})
	}
}

func TestGetDeviceState_Error(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type env struct {
			RequestID string `json:"requestId"`
			Code      int    `json:"code"`
			Msg       string `json:"msg"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(env{Code: 401, Msg: "Unauthorized"}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))

	_, err := client.GetDeviceState(context.Background(), "H6008", "AA:BB")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*govee.APIError)
	if !ok {
		t.Fatalf("expected *govee.APIError, got %T: %v", err, err)
	}
	if apiErr.Code != http.StatusUnauthorized {
		t.Errorf("Code: got %d, want %d", apiErr.Code, http.StatusUnauthorized)
	}
	_ = fmt.Sprintf("error: %v", apiErr)
}
