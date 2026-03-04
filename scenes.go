package govee

import (
	"context"
	"encoding/json"
	"fmt"
)

// SceneOption is a named dynamic light scene available for a device.
// The Value contains the id and paramId required to activate it via SetLightScene.
type SceneOption struct {
	Name  string          `json:"name"`
	Value LightSceneValue `json:"value"`
}

// DIYSceneOption is a user-created DIY scene available for a device.
// Activate it with SetDIYScene using the integer Value.
type DIYSceneOption struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// scenesPayload is the POST body shared by /device/scenes and /device/diy-scenes.
type scenesPayload struct {
	SKU      string `json:"sku"`
	DeviceID string `json:"device"`
}

// scenesResponse holds the capability list returned by the scenes endpoints.
type scenesResponse struct {
	SKU          string       `json:"sku"`
	DeviceID     string       `json:"device"`
	Capabilities []Capability `json:"capabilities"`
}

// GetScenes returns the dynamic light scenes available for a device. These scenes
// have compound LightSceneValue values (id + paramId) and are separate from the
// static scenes embedded in device capabilities returned by GetDevices.
func (c *Client) GetScenes(ctx context.Context, sku, deviceID string) ([]SceneOption, error) {
	data, err := c.doPost(ctx, "/device/scenes", scenesPayload{
		SKU:      sku,
		DeviceID: deviceID,
	})
	if err != nil {
		return nil, err
	}

	var resp scenesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("govee: failed to decode scenes response: %w", err)
	}

	for _, cap := range resp.Capabilities {
		if cap.Type == CapabilityDynamicScene && cap.Instance == "lightScene" {
			var params EnumParameters
			if err := json.Unmarshal(cap.Parameters, &params); err != nil {
				return nil, fmt.Errorf("govee: failed to decode scene parameters: %w", err)
			}
			scenes := make([]SceneOption, 0, len(params.Options))
			for _, opt := range params.Options {
				var val LightSceneValue
				if err := json.Unmarshal(opt.Value, &val); err != nil {
					return nil, fmt.Errorf("govee: failed to decode scene value for %q: %w", opt.Name, err)
				}
				scenes = append(scenes, SceneOption{Name: opt.Name, Value: val})
			}
			return scenes, nil
		}
	}
	return nil, nil
}

// GetDIYScenes returns the user-created DIY scenes available for a device.
// Activate a scene with SetDIYScene using the integer Value from each DIYSceneOption.
func (c *Client) GetDIYScenes(ctx context.Context, sku, deviceID string) ([]DIYSceneOption, error) {
	data, err := c.doPost(ctx, "/device/diy-scenes", scenesPayload{
		SKU:      sku,
		DeviceID: deviceID,
	})
	if err != nil {
		return nil, err
	}

	var resp scenesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("govee: failed to decode DIY scenes response: %w", err)
	}

	for _, cap := range resp.Capabilities {
		if cap.Instance == "diyScene" {
			var params EnumParameters
			if err := json.Unmarshal(cap.Parameters, &params); err != nil {
				return nil, fmt.Errorf("govee: failed to decode DIY scene parameters: %w", err)
			}
			scenes := make([]DIYSceneOption, 0, len(params.Options))
			for _, opt := range params.Options {
				var val int
				if err := json.Unmarshal(opt.Value, &val); err != nil {
					return nil, fmt.Errorf("govee: failed to decode DIY scene value for %q: %w", opt.Name, err)
				}
				scenes = append(scenes, DIYSceneOption{Name: opt.Name, Value: val})
			}
			return scenes, nil
		}
	}
	return nil, nil
}
