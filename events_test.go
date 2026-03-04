package govee_test

import (
	"encoding/json"
	"testing"

	govee "github.com/DTCurrie/govee-go"
)

func TestNewEventClient(t *testing.T) {
	t.Parallel()

	handler := func(govee.DeviceEvent) {}
	ec := govee.NewEventClient("test-key", handler)
	if ec == nil {
		t.Fatal("NewEventClient returned nil")
	}
}

func TestNewEventClient_WithBroker(t *testing.T) {
	t.Parallel()

	handler := func(govee.DeviceEvent) {}
	ec := govee.NewEventClient("test-key", handler, govee.WithMQTTBroker("mqtt://localhost:1883"))
	if ec == nil {
		t.Fatal("NewEventClient returned nil")
	}
}

func TestEventClient_Close_NotConnected(t *testing.T) {
	t.Parallel()

	handler := func(govee.DeviceEvent) {}
	ec := govee.NewEventClient("test-key", handler)
	if err := ec.Close(); err != nil {
		t.Errorf("Close() on unconnected client: %v", err)
	}
}

func TestDeviceEvent_JSONParsing(t *testing.T) {
	t.Parallel()

	payload := `{
		"sku": "H6008",
		"device": "AA:BB:CC:DD",
		"deviceName": "Bedroom Light",
		"capabilities": [
			{"type": "devices.capabilities.on_off", "instance": "powerSwitch", "state": 1},
			{"type": "devices.capabilities.range", "instance": "brightness", "state": 75}
		]
	}`

	var event govee.DeviceEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Unmarshal DeviceEvent: %v", err)
	}
	if event.SKU != "H6008" {
		t.Errorf("SKU: got %q, want H6008", event.SKU)
	}
	if event.DeviceID != "AA:BB:CC:DD" {
		t.Errorf("DeviceID: got %q, want AA:BB:CC:DD", event.DeviceID)
	}
	if event.DeviceName != "Bedroom Light" {
		t.Errorf("DeviceName: got %q, want Bedroom Light", event.DeviceName)
	}
	if len(event.Capabilities) != 2 {
		t.Fatalf("capabilities count: got %d, want 2", len(event.Capabilities))
	}
	if event.Capabilities[0].Type != govee.CapabilityOnOff {
		t.Errorf("capabilities[0].Type: got %q, want %q", event.Capabilities[0].Type, govee.CapabilityOnOff)
	}
	if event.Capabilities[0].Instance != "powerSwitch" {
		t.Errorf("capabilities[0].Instance: got %q, want powerSwitch", event.Capabilities[0].Instance)
	}
	if string(event.Capabilities[0].State) != "1" {
		t.Errorf("capabilities[0].State: got %q, want 1", event.Capabilities[0].State)
	}
}

func TestDeviceEvent_EmptyCapabilities(t *testing.T) {
	t.Parallel()

	payload := `{"sku": "H6008", "device": "AA:BB", "deviceName": "", "capabilities": []}`

	var event govee.DeviceEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(event.Capabilities) != 0 {
		t.Errorf("expected 0 capabilities, got %d", len(event.Capabilities))
	}
}

func TestEventCapability_JSONParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		json      string
		wantType  string
		wantInst  string
		wantState string
	}{
		{
			"integer state",
			`{"type":"devices.capabilities.on_off","instance":"powerSwitch","state":1}`,
			govee.CapabilityOnOff, "powerSwitch", "1",
		},
		{
			"object state",
			`{"type":"devices.capabilities.color_setting","instance":"colorRgb","state":{"r":255,"g":0,"b":0}}`,
			govee.CapabilityColorSetting, "colorRgb", `{"r":255,"g":0,"b":0}`,
		},
		{
			"boolean state",
			`{"type":"devices.capabilities.online","instance":"online","state":true}`,
			govee.CapabilityOnline, "online", "true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cap govee.EventCapability
			if err := json.Unmarshal([]byte(tc.json), &cap); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if cap.Type != tc.wantType {
				t.Errorf("Type: got %q, want %q", cap.Type, tc.wantType)
			}
			if cap.Instance != tc.wantInst {
				t.Errorf("Instance: got %q, want %q", cap.Instance, tc.wantInst)
			}
			if string(cap.State) != tc.wantState {
				t.Errorf("State: got %q, want %q", cap.State, tc.wantState)
			}
		})
	}
}
