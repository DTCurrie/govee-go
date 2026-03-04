package govee_test

import (
	"context"
	"net/http"
	"testing"

	govee "github.com/DTCurrie/govee-go"
)

// capabilityFromBody extracts the nested "capability" object from a decoded POST body.
func capabilityFromBody(t *testing.T, body map[string]interface{}) map[string]interface{} {
	t.Helper()
	p := postPayload(t, body)
	cap, ok := p["capability"].(map[string]interface{})
	if !ok {
		t.Errorf("capability field not found or not an object; payload = %v", p)
	}
	return cap
}

func makeControlServer(t *testing.T, captureCap *map[string]interface{}) *govee.Client {
	t.Helper()
	return newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/control" {
			t.Errorf("expected /device/control, got %s", r.URL.Path)
		}
		body := decodePostBody(t, r)
		*captureCap = capabilityFromBody(t, body)
		respondPostJSON(t, w, struct{}{})
	}))
}

func TestControlDevice(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	err := client.ControlDevice(context.Background(), "H6008", "AA:BB", govee.CapabilityCommand{
		Type:     govee.CapabilityOnOff,
		Instance: "powerSwitch",
		Value:    1,
	})
	if err != nil {
		t.Fatalf("ControlDevice: %v", err)
	}
	if cap["type"] != govee.CapabilityOnOff {
		t.Errorf("type: got %v, want %v", cap["type"], govee.CapabilityOnOff)
	}
	if cap["instance"] != "powerSwitch" {
		t.Errorf("instance: got %v, want powerSwitch", cap["instance"])
	}
}

func TestControlDevice_Error(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondAPIError(t, w, http.StatusTooManyRequests, "rate limit")
	}))

	err := client.TurnOn(context.Background(), "H6008", "AA:BB")
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
}

func TestTurnOn(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	if err := client.TurnOn(context.Background(), "H6008", "AA:BB"); err != nil {
		t.Fatalf("TurnOn: %v", err)
	}
	if cap["type"] != govee.CapabilityOnOff {
		t.Errorf("type: got %v, want %v", cap["type"], govee.CapabilityOnOff)
	}
	if cap["instance"] != "powerSwitch" {
		t.Errorf("instance: got %v, want powerSwitch", cap["instance"])
	}
	if cap["value"].(float64) != 1 {
		t.Errorf("value: got %v, want 1", cap["value"])
	}
}

func TestTurnOff(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	if err := client.TurnOff(context.Background(), "H6008", "AA:BB"); err != nil {
		t.Fatalf("TurnOff: %v", err)
	}
	if cap["value"].(float64) != 0 {
		t.Errorf("value: got %v, want 0", cap["value"])
	}
}

func TestSetBrightness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent int
		wantErr bool
	}{
		{"minimum valid", 1, false},
		{"maximum valid", 100, false},
		{"midpoint", 50, false},
		{"zero is invalid", 0, true},
		{"negative is invalid", -5, true},
		{"over 100 is invalid", 101, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cap map[string]interface{}
			client := makeControlServer(t, &cap)

			err := client.SetBrightness(context.Background(), "H6008", "AA:BB", tc.percent)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for percent=%d, got nil", tc.percent)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetBrightness(%d): %v", tc.percent, err)
			}
			if cap["instance"] != "brightness" {
				t.Errorf("instance: got %v, want brightness", cap["instance"])
			}
			if cap["value"].(float64) != float64(tc.percent) {
				t.Errorf("value: got %v, want %d", cap["value"], tc.percent)
			}
		})
	}
}

func TestSetColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		r, g, b int
		wantErr bool
	}{
		{"valid red", 255, 0, 0, false},
		{"valid white", 255, 255, 255, false},
		{"valid off", 0, 0, 0, false},
		{"r out of range", 256, 0, 0, true},
		{"g out of range", 0, 256, 0, true},
		{"b out of range", 0, 0, 256, true},
		{"r negative", -1, 0, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cap map[string]interface{}
			client := makeControlServer(t, &cap)

			err := client.SetColor(context.Background(), "H6008", "AA:BB", tc.r, tc.g, tc.b)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for r=%d g=%d b=%d, got nil", tc.r, tc.g, tc.b)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetColor: %v", err)
			}
			want := float64((tc.r << 16) | (tc.g << 8) | tc.b)
			if cap["value"].(float64) != want {
				t.Errorf("value: got %v, want %v", cap["value"], want)
			}
		})
	}
}

func TestSetColorTemp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		kelvin  int
		wantErr bool
	}{
		{"valid 4000K", 4000, false},
		{"valid 6500K", 6500, false},
		{"zero is invalid", 0, true},
		{"negative is invalid", -100, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cap map[string]interface{}
			client := makeControlServer(t, &cap)

			err := client.SetColorTemp(context.Background(), "H6008", "AA:BB", tc.kelvin)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for kelvin=%d, got nil", tc.kelvin)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetColorTemp(%d): %v", tc.kelvin, err)
			}
			if cap["instance"] != "colorTemperatureK" {
				t.Errorf("instance: got %v, want colorTemperatureK", cap["instance"])
			}
		})
	}
}

func TestSetToggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		instance  string
		on        bool
		wantValue float64
	}{
		{"gradient on", "gradientToggle", true, 1},
		{"gradient off", "gradientToggle", false, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cap map[string]interface{}
			client := makeControlServer(t, &cap)

			if err := client.SetToggle(context.Background(), "H6008", "AA:BB", tc.instance, tc.on); err != nil {
				t.Fatalf("SetToggle: %v", err)
			}
			if cap["instance"] != tc.instance {
				t.Errorf("instance: got %v, want %v", cap["instance"], tc.instance)
			}
			if cap["value"].(float64) != tc.wantValue {
				t.Errorf("value: got %v, want %v", cap["value"], tc.wantValue)
			}
		})
	}
}

func TestSetLightScene(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	val := govee.LightSceneValue{ID: 42, ParamID: 5}
	if err := client.SetLightScene(context.Background(), "H6008", "AA:BB", val); err != nil {
		t.Fatalf("SetLightScene: %v", err)
	}
	if cap["instance"] != "lightScene" {
		t.Errorf("instance: got %v, want lightScene", cap["instance"])
	}
}

func TestSetDIYScene(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	if err := client.SetDIYScene(context.Background(), "H6008", "AA:BB", 7); err != nil {
		t.Fatalf("SetDIYScene: %v", err)
	}
	if cap["instance"] != "diyScene" {
		t.Errorf("instance: got %v, want diyScene", cap["instance"])
	}
	if cap["value"].(float64) != 7 {
		t.Errorf("value: got %v, want 7", cap["value"])
	}
}

func TestSetSnapshot(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	if err := client.SetSnapshot(context.Background(), "H6008", "AA:BB", 3); err != nil {
		t.Fatalf("SetSnapshot: %v", err)
	}
	if cap["instance"] != "snapshot" {
		t.Errorf("instance: got %v, want snapshot", cap["instance"])
	}
}

func TestSetMusicMode(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	mode := govee.MusicModeValue{MusicMode: 2, Sensitivity: 80}
	if err := client.SetMusicMode(context.Background(), "H6008", "AA:BB", mode); err != nil {
		t.Fatalf("SetMusicMode: %v", err)
	}
	if cap["instance"] != "musicMode" {
		t.Errorf("instance: got %v, want musicMode", cap["instance"])
	}
}

func TestSetWorkMode(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	mode := govee.WorkModeValue{WorkMode: 1, ModeValue: 3}
	if err := client.SetWorkMode(context.Background(), "H6008", "AA:BB", mode); err != nil {
		t.Fatalf("SetWorkMode: %v", err)
	}
	if cap["instance"] != "workMode" {
		t.Errorf("instance: got %v, want workMode", cap["instance"])
	}
}

func TestSetSegmentColor(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	segments := []int{0, 1, 2}
	if err := client.SetSegmentColor(context.Background(), "H6008", "AA:BB", segments, 0xFF0000); err != nil {
		t.Fatalf("SetSegmentColor: %v", err)
	}
	if cap["instance"] != "segmentedColorRgb" {
		t.Errorf("instance: got %v, want segmentedColorRgb", cap["instance"])
	}
}

func TestSetSegmentBrightness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		brightness int
		wantErr    bool
	}{
		{"valid 50", 50, false},
		{"valid 0", 0, false},
		{"valid 100", 100, false},
		{"over 100", 101, true},
		{"negative", -1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cap map[string]interface{}
			client := makeControlServer(t, &cap)

			err := client.SetSegmentBrightness(context.Background(), "H6008", "AA:BB", []int{0}, tc.brightness)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for brightness=%d, got nil", tc.brightness)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetSegmentBrightness(%d): %v", tc.brightness, err)
			}
			if cap["instance"] != "segmentedBrightness" {
				t.Errorf("instance: got %v, want segmentedBrightness", cap["instance"])
			}
		})
	}
}

func TestSetTemperature(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	temp := govee.TemperatureValue{Temperature: 100, Unit: "C"}
	if err := client.SetTemperature(context.Background(), "H6008", "AA:BB", temp); err != nil {
		t.Fatalf("SetTemperature: %v", err)
	}
	if cap["instance"] != "targetTemperature" {
		t.Errorf("instance: got %v, want targetTemperature", cap["instance"])
	}
}

func TestSetMode(t *testing.T) {
	t.Parallel()

	var cap map[string]interface{}
	client := makeControlServer(t, &cap)

	if err := client.SetMode(context.Background(), "H6008", "AA:BB", "nightlightScene", 5); err != nil {
		t.Fatalf("SetMode: %v", err)
	}
	if cap["instance"] != "nightlightScene" {
		t.Errorf("instance: got %v, want nightlightScene", cap["instance"])
	}
	if cap["value"].(float64) != 5 {
		t.Errorf("value: got %v, want 5", cap["value"])
	}
}
