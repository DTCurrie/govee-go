package govee_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	govee "github.com/DTCurrie/govee-go"
)

func TestGetScenes(t *testing.T) {
	t.Parallel()

	// The API returns a payload with capabilities; the lightScene capability holds
	// an ENUM of options, each with a {id, paramId} value.
	sceneOptions := []map[string]interface{}{
		{"name": "Rainbow", "value": map[string]interface{}{"id": 1, "paramId": 0}},
		{"name": "Ocean", "value": map[string]interface{}{"id": 2, "paramId": 1}},
	}
	payload := map[string]interface{}{
		"sku":    "H6008",
		"device": "AA:BB",
		"capabilities": []map[string]interface{}{
			{
				"type":     govee.CapabilityDynamicScene,
				"instance": "lightScene",
				"parameters": map[string]interface{}{
					"dataType": "ENUM",
					"options":  sceneOptions,
				},
			},
		},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/scenes" {
			t.Errorf("expected /device/scenes, got %s", r.URL.Path)
		}
		body := decodePostBody(t, r)
		p := postPayload(t, body)
		if p["sku"] != "H6008" {
			t.Errorf("sku: got %v, want H6008", p["sku"])
		}
		respondPostJSON(t, w, payload)
	}))

	scenes, err := client.GetScenes(context.Background(), "H6008", "AA:BB")
	if err != nil {
		t.Fatalf("GetScenes: %v", err)
	}
	if len(scenes) != 2 {
		t.Fatalf("scene count: got %d, want 2", len(scenes))
	}
	if scenes[0].Name != "Rainbow" {
		t.Errorf("scenes[0].Name: got %q, want Rainbow", scenes[0].Name)
	}
	if scenes[0].Value.ID != 1 {
		t.Errorf("scenes[0].Value.ID: got %d, want 1", scenes[0].Value.ID)
	}
	if scenes[1].Name != "Ocean" {
		t.Errorf("scenes[1].Name: got %q, want Ocean", scenes[1].Name)
	}
	if scenes[1].Value.ParamID != 1 {
		t.Errorf("scenes[1].Value.ParamID: got %d, want 1", scenes[1].Value.ParamID)
	}
}

func TestGetScenes_NoCapability(t *testing.T) {
	t.Parallel()

	payload := map[string]interface{}{
		"sku":          "H6008",
		"device":       "AA:BB",
		"capabilities": []interface{}{},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondPostJSON(t, w, payload)
	}))

	scenes, err := client.GetScenes(context.Background(), "H6008", "AA:BB")
	if err != nil {
		t.Fatalf("GetScenes: %v", err)
	}
	if scenes != nil {
		t.Errorf("expected nil scenes for no capability, got %v", scenes)
	}
}

func TestGetScenes_Error(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondAPIError(t, w, http.StatusUnauthorized, "Unauthorized")
	}))

	_, err := client.GetScenes(context.Background(), "H6008", "AA:BB")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDIYScenes(t *testing.T) {
	t.Parallel()

	diyOptions := []map[string]interface{}{
		{"name": "My Scene 1", "value": json.RawMessage(`10`)},
		{"name": "My Scene 2", "value": json.RawMessage(`20`)},
	}
	payload := map[string]interface{}{
		"sku":    "H6008",
		"device": "AA:BB",
		"capabilities": []map[string]interface{}{
			{
				"type":     govee.CapabilityDynamicScene,
				"instance": "diyScene",
				"parameters": map[string]interface{}{
					"dataType": "ENUM",
					"options":  diyOptions,
				},
			},
		},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device/diy-scenes" {
			t.Errorf("expected /device/diy-scenes, got %s", r.URL.Path)
		}
		respondPostJSON(t, w, payload)
	}))

	scenes, err := client.GetDIYScenes(context.Background(), "H6008", "AA:BB")
	if err != nil {
		t.Fatalf("GetDIYScenes: %v", err)
	}
	if len(scenes) != 2 {
		t.Fatalf("scene count: got %d, want 2", len(scenes))
	}
	if scenes[0].Name != "My Scene 1" {
		t.Errorf("scenes[0].Name: got %q, want My Scene 1", scenes[0].Name)
	}
	if scenes[0].Value != 10 {
		t.Errorf("scenes[0].Value: got %d, want 10", scenes[0].Value)
	}
	if scenes[1].Value != 20 {
		t.Errorf("scenes[1].Value: got %d, want 20", scenes[1].Value)
	}
}

func TestGetDIYScenes_NoCapability(t *testing.T) {
	t.Parallel()

	payload := map[string]interface{}{
		"sku":          "H6008",
		"device":       "AA:BB",
		"capabilities": []interface{}{},
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondPostJSON(t, w, payload)
	}))

	scenes, err := client.GetDIYScenes(context.Background(), "H6008", "AA:BB")
	if err != nil {
		t.Fatalf("GetDIYScenes: %v", err)
	}
	if scenes != nil {
		t.Errorf("expected nil scenes for no capability, got %v", scenes)
	}
}
