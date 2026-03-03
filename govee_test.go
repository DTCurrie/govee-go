package govee

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at a test server and returns both.
func newTestClient(t *testing.T, mux *http.ServeMux) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := New("test-api-key", WithBaseURL(srv.URL))
	return client, srv
}

// respondJSON writes a Govee-style JSON envelope response.
func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	resp := map[string]interface{}{
		"code":    200,
		"message": "Success",
		"data":    json.RawMessage(dataBytes),
	}
	if code != 200 {
		resp["code"] = code
		resp["message"] = http.StatusText(code)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func TestGetDevices(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Govee-API-Key") != "test-api-key" {
			t.Errorf("missing or wrong API key header")
		}
		respondJSON(w, 200, map[string]interface{}{
			"devices": []map[string]interface{}{
				{
					"device":       "AA:BB:CC:DD:EE:FF:00:11",
					"model":        "H6159",
					"deviceName":   "Living Room Light",
					"controllable": true,
					"retrievable":  true,
					"supportCmds":  []string{"turn", "brightness", "color", "colorTem"},
					"properties": map[string]interface{}{
						"colorTem": map[string]interface{}{
							"range": map[string]interface{}{"min": 2000, "max": 9000},
						},
					},
				},
				{
					"device":       "11:22:33:44:55:66:77:88",
					"model":        "H5081",
					"deviceName":   "Smart Plug",
					"controllable": true,
					"retrievable":  false,
					"supportCmds":  []string{"turn"},
					"properties":   map[string]interface{}{},
				},
			},
		})
	})

	client, _ := newTestClient(t, mux)
	devices, err := client.GetDevices(context.Background())
	if err != nil {
		t.Fatalf("GetDevices: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}

	d0 := devices[0]
	if d0.DeviceID != "AA:BB:CC:DD:EE:FF:00:11" {
		t.Errorf("device[0] ID = %q", d0.DeviceID)
	}
	if d0.Model != "H6159" {
		t.Errorf("device[0] Model = %q", d0.Model)
	}
	if d0.DeviceName != "Living Room Light" {
		t.Errorf("device[0] Name = %q", d0.DeviceName)
	}
	if !d0.Controllable || !d0.Retrievable {
		t.Error("device[0] should be controllable and retrievable")
	}
	if d0.Properties.ColorTemp == nil {
		t.Fatal("device[0] ColorTemp should not be nil")
	}
	if d0.Properties.ColorTemp.Min != 2000 || d0.Properties.ColorTemp.Max != 9000 {
		t.Errorf("device[0] ColorTemp range = %d-%d", d0.Properties.ColorTemp.Min, d0.Properties.ColorTemp.Max)
	}

	d1 := devices[1]
	if d1.Properties.ColorTemp != nil {
		t.Error("device[1] should have nil ColorTemp")
	}
	if d1.Retrievable {
		t.Error("device[1] should not be retrievable")
	}
}

func TestGetDevices_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    403,
			"message": "Unauthorized",
			"data":    nil,
		})
	})

	client, _ := newTestClient(t, mux)
	_, err := client.GetDevices(context.Background())
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != 403 {
		t.Errorf("expected code 403, got %d", apiErr.Code)
	}
}

func TestGetDeviceState(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/devices/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("device") != "AA:BB:CC" {
			t.Errorf("missing device query param")
		}
		if r.URL.Query().Get("model") != "H6159" {
			t.Errorf("missing model query param")
		}
		respondJSON(w, 200, map[string]interface{}{
			"device": "AA:BB:CC",
			"model":  "H6159",
			"properties": []interface{}{
				map[string]interface{}{"online": "true"},
				map[string]interface{}{"powerState": "on"},
				map[string]interface{}{"brightness": 80},
				map[string]interface{}{"color": map[string]interface{}{"r": 255, "g": 128, "b": 0}},
				map[string]interface{}{"colorTem": 5000},
			},
		})
	})

	client, _ := newTestClient(t, mux)
	state, err := client.GetDeviceState(context.Background(), "AA:BB:CC", "H6159")
	if err != nil {
		t.Fatalf("GetDeviceState: %v", err)
	}
	if !state.Online {
		t.Error("expected Online=true")
	}
	if state.PowerState != "on" {
		t.Errorf("PowerState = %q", state.PowerState)
	}
	if state.Brightness != 80 {
		t.Errorf("Brightness = %d", state.Brightness)
	}
	if state.Color == nil {
		t.Fatal("Color should not be nil")
	}
	if state.Color.R != 255 || state.Color.G != 128 || state.Color.B != 0 {
		t.Errorf("Color = %+v", state.Color)
	}
	if state.ColorTemp != 5000 {
		t.Errorf("ColorTemp = %d", state.ColorTemp)
	}
}

func TestTurnOn(t *testing.T) {
	testControl(t, func(client *Client) error {
		return client.TurnOn(context.Background(), "AA:BB:CC", "H6159")
	}, "turn", "on")
}

func TestTurnOff(t *testing.T) {
	testControl(t, func(client *Client) error {
		return client.TurnOff(context.Background(), "AA:BB:CC", "H6159")
	}, "turn", "off")
}

func TestSetBrightness(t *testing.T) {
	testControl(t, func(client *Client) error {
		return client.SetBrightness(context.Background(), "AA:BB:CC", "H6159", 75)
	}, "brightness", float64(75))
}

func TestSetColor(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/devices/control", func(w http.ResponseWriter, r *http.Request) {
		var req controlRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Cmd.Name != "color" {
			t.Errorf("cmd.name = %q, want %q", req.Cmd.Name, "color")
		}
		// Value is decoded as map[string]interface{} by JSON
		colorMap, ok := req.Cmd.Value.(map[string]interface{})
		if !ok {
			t.Fatalf("cmd.value type = %T", req.Cmd.Value)
		}
		if colorMap["r"] != float64(255) || colorMap["g"] != float64(0) || colorMap["b"] != float64(128) {
			t.Errorf("color value = %v", colorMap)
		}
		respondJSON(w, 200, map[string]interface{}{})
	})

	client, _ := newTestClient(t, mux)
	if err := client.SetColor(context.Background(), "AA:BB:CC", "H6159", 255, 0, 128); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
}

func TestSetColorTemp(t *testing.T) {
	testControl(t, func(client *Client) error {
		return client.SetColorTemp(context.Background(), "AA:BB:CC", "H6159", 6500)
	}, "colorTem", float64(6500))
}

func TestSetBrightnessValidation(t *testing.T) {
	client := New("key")
	if err := client.SetBrightness(context.Background(), "d", "m", -1); err == nil {
		t.Error("expected error for brightness -1")
	}
	if err := client.SetBrightness(context.Background(), "d", "m", 101); err == nil {
		t.Error("expected error for brightness 101")
	}
}

func TestSetColorValidation(t *testing.T) {
	client := New("key")
	if err := client.SetColor(context.Background(), "d", "m", -1, 0, 0); err == nil {
		t.Error("expected error for r=-1")
	}
	if err := client.SetColor(context.Background(), "d", "m", 0, 256, 0); err == nil {
		t.Error("expected error for g=256")
	}
	if err := client.SetColor(context.Background(), "d", "m", 0, 0, 300); err == nil {
		t.Error("expected error for b=300")
	}
}

func TestSetColorTempValidation(t *testing.T) {
	client := New("key")
	if err := client.SetColorTemp(context.Background(), "d", "m", 0); err == nil {
		t.Error("expected error for kelvin=0")
	}
	if err := client.SetColorTemp(context.Background(), "d", "m", -100); err == nil {
		t.Error("expected error for kelvin=-100")
	}
}

func TestSupportsCmdHelper(t *testing.T) {
	d := Device{SupportCmds: []string{"turn", "brightness", "color"}}
	if !d.SupportsCmd("color") {
		t.Error("expected SupportsCmd(color) = true")
	}
	if d.SupportsCmd("colorTemp") {
		t.Error("expected SupportsCmd(colorTemp) = false")
	}
}

// testControl is a helper that verifies a control call sends the right cmd name/value.
func testControl(t *testing.T, call func(*Client) error, wantName string, wantValue interface{}) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/devices/control", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		var req controlRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Device != "AA:BB:CC" {
			t.Errorf("device = %q", req.Device)
		}
		if req.Model != "H6159" {
			t.Errorf("model = %q", req.Model)
		}
		if req.Cmd.Name != wantName {
			t.Errorf("cmd.name = %q, want %q", req.Cmd.Name, wantName)
		}
		if req.Cmd.Value != wantValue {
			t.Errorf("cmd.value = %v (%T), want %v (%T)", req.Cmd.Value, req.Cmd.Value, wantValue, wantValue)
		}
		respondJSON(w, 200, map[string]interface{}{})
	})

	client, _ := newTestClient(t, mux)
	if err := call(client); err != nil {
		t.Fatalf("control call failed: %v", err)
	}
}
