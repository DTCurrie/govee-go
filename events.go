package govee

import (
	"context"
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const defaultMQTTBroker = "mqtts://mqtt.openapi.govee.com:8883"

// EventCapability is a single capability state update delivered inside a DeviceEvent.
type EventCapability struct {
	Type     string          `json:"type"`
	Instance string          `json:"instance"`
	State    json.RawMessage `json:"state"`
}

// DeviceEvent is a device state notification received from the Govee MQTT broker.
type DeviceEvent struct {
	SKU          string            `json:"sku"`
	DeviceID     string            `json:"device"`
	DeviceName   string            `json:"deviceName"`
	Capabilities []EventCapability `json:"capabilities"`
}

// EventHandler is a callback invoked for each DeviceEvent received over MQTT.
type EventHandler func(event DeviceEvent)

// EventOption is a functional option for configuring an EventClient.
type EventOption func(*EventClient)

// WithMQTTBroker overrides the MQTT broker URL used by an EventClient.
// The default is mqtts://mqtt.openapi.govee.com:8883.
func WithMQTTBroker(url string) EventOption {
	return func(e *EventClient) {
		e.brokerURL = url
	}
}

// EventClient subscribes to real-time device events from the Govee MQTT broker.
// Create one with NewEventClient, then call Connect to begin receiving events.
type EventClient struct {
	apiKey     string
	brokerURL  string
	handler    EventHandler
	mqttClient mqtt.Client
}

// NewEventClient creates a new EventClient that delivers events to handler.
func NewEventClient(apiKey string, handler EventHandler, opts ...EventOption) *EventClient {
	e := &EventClient{
		apiKey:    apiKey,
		brokerURL: defaultMQTTBroker,
		handler:   handler,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Connect establishes the MQTT connection and starts delivering events to the handler.
// It blocks until the initial connection is established or ctx is cancelled.
func (e *EventClient) Connect(ctx context.Context) error {
	topic := "GA/" + e.apiKey

	opts := mqtt.NewClientOptions().
		AddBroker(e.brokerURL).
		SetUsername(e.apiKey).
		SetPassword(e.apiKey).
		SetAutoReconnect(true).
		SetOnConnectHandler(func(c mqtt.Client) {
			c.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) { //nolint:errcheck
				var event DeviceEvent
				if err := json.Unmarshal(msg.Payload(), &event); err != nil {
					return
				}
				e.handler(event)
			})
		})

	e.mqttClient = mqtt.NewClient(opts)
	token := e.mqttClient.Connect()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-token.Done():
		if err := token.Error(); err != nil {
			return fmt.Errorf("govee: MQTT connect failed: %w", err)
		}
	}
	return nil
}

// Close disconnects the MQTT client and stops event delivery.
func (e *EventClient) Close() error {
	if e.mqttClient != nil && e.mqttClient.IsConnected() {
		e.mqttClient.Disconnect(250)
	}
	return nil
}
