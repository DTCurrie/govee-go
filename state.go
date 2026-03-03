package govee

import (
	"context"
	"encoding/json"
	"fmt"
)

// statePropertyJSON represents one entry in the properties array returned by
// GET /v1/devices/state. Each entry is a single-key object, e.g.:
//
//	{"online": "true"}
//	{"powerState": "on"}
//	{"brightness": 80}
//	{"color": {"r": 255, "g": 0, "b": 0}}
//	{"colorTemp": 5000}
type statePropertyJSON map[string]json.RawMessage

// stateResponse is the shape of the data field from GET /v1/devices/state.
type stateResponse struct {
	Device     string              `json:"device"`
	Model      string              `json:"model"`
	Properties []statePropertyJSON `json:"properties"`
}

// GetDeviceState queries the current state of a device.
// Only devices with Retrievable=true return useful state.
// Note: the "online" field is cached on Govee's side and may be stale.
func (c *Client) GetDeviceState(ctx context.Context, deviceID, model string) (*DeviceState, error) {
	data, err := c.doGet(ctx, "/v1/devices/state", map[string]string{
		"device": deviceID,
		"model":  model,
	})
	if err != nil {
		return nil, err
	}

	var resp stateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("govee: failed to decode device state: %w", err)
	}

	state := &DeviceState{}
	for _, prop := range resp.Properties {
		for key, raw := range prop {
			switch key {
			case "online":
				var v string
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("govee: failed to parse state property %q: %w", key, err)
				}
				state.Online = v == "true"
			case "powerState":
				var v string
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("govee: failed to parse state property %q: %w", key, err)
				}
				state.PowerState = v
			case "brightness":
				var v int
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("govee: failed to parse state property %q: %w", key, err)
				}
				state.Brightness = v
			case "color":
				var v Color
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("govee: failed to parse state property %q: %w", key, err)
				}
				state.Color = &v
			case "colorTem":
				var v int
				if err := json.Unmarshal(raw, &v); err != nil {
					return nil, fmt.Errorf("govee: failed to parse state property %q: %w", key, err)
				}
				state.ColorTemp = v
			}
		}
	}
	return state, nil
}
