package govee_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	govee "github.com/DTCurrie/govee-go"
)

func ExampleClient_GetDevices() {
	// Normally you would use govee.New("your-api-key").
	// Here we use a test server so the example is self-contained.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		devices := []govee.Device{
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
		}
		type envelope struct {
			Code    int            `json:"code"`
			Message string         `json:"message"`
			Data    []govee.Device `json:"data"`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envelope{Code: 200, Message: "success", Data: devices})
	}))
	defer srv.Close()

	client := govee.New("my-api-key", govee.WithBaseURL(srv.URL))

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, d := range devices {
		fmt.Printf("%s (%s)\n", d.DeviceName, d.SKU)
	}
	// Output:
	// Bedroom Light (H6008)
}

func ExampleClient_TurnOn() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type envelope struct {
			RequestID string      `json:"requestId"`
			Code      int         `json:"code"`
			Msg       string      `json:"msg"`
			Payload   interface{} `json:"payload"`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envelope{Code: 200, Msg: "success", Payload: struct{}{}})
	}))
	defer srv.Close()

	client := govee.New("my-api-key", govee.WithBaseURL(srv.URL))

	err := client.TurnOn(context.Background(), "H6008", "AA:BB:CC:DD")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("turned on")
	// Output:
	// turned on
}

func ExampleClient_SetBrightness() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type envelope struct {
			RequestID string      `json:"requestId"`
			Code      int         `json:"code"`
			Msg       string      `json:"msg"`
			Payload   interface{} `json:"payload"`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envelope{Code: 200, Msg: "success", Payload: struct{}{}})
	}))
	defer srv.Close()

	client := govee.New("my-api-key", govee.WithBaseURL(srv.URL))

	err := client.SetBrightness(context.Background(), "H6008", "AA:BB:CC:DD", 80)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("brightness set")
	// Output:
	// brightness set
}

func ExampleClient_SetLightScene() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type envelope struct {
			RequestID string      `json:"requestId"`
			Code      int         `json:"code"`
			Msg       string      `json:"msg"`
			Payload   interface{} `json:"payload"`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envelope{Code: 200, Msg: "success", Payload: struct{}{}})
	}))
	defer srv.Close()

	client := govee.New("my-api-key", govee.WithBaseURL(srv.URL))

	// Obtain scene values from GetScenes, then activate with SetLightScene.
	scene := govee.LightSceneValue{ID: 2, ParamID: 0}
	err := client.SetLightScene(context.Background(), "H6008", "AA:BB:CC:DD", scene)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("scene activated")
	// Output:
	// scene activated
}

func ExampleNewEventClient() {
	handler := func(event govee.DeviceEvent) {
		fmt.Printf("event from %s: %d capability updates\n", event.SKU, len(event.Capabilities))
	}

	ec := govee.NewEventClient("my-api-key", handler,
		govee.WithMQTTBroker("mqtt://localhost:1883"),
	)
	// ec.Connect(ctx) would establish the MQTT connection.
	_ = ec
}
