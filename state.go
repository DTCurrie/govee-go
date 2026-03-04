package govee

import (
	"context"
	"encoding/json"
	"fmt"
)

// DeviceStateResponse holds the current state of a device expressed as a list of
// capabilities, each carrying their current value in the State field.
type DeviceStateResponse struct {
	SKU          string       `json:"sku"`
	DeviceID     string       `json:"device"`
	Capabilities []Capability `json:"capabilities"`
}

// FindState returns the capability with the given type and instance, or nil if not found.
func (r *DeviceStateResponse) FindState(capType, instance string) *Capability {
	for i := range r.Capabilities {
		if r.Capabilities[i].Type == capType && r.Capabilities[i].Instance == instance {
			return &r.Capabilities[i]
		}
	}
	return nil
}

// IsOnline reports whether the device is currently online.
func (r *DeviceStateResponse) IsOnline() bool {
	cap := r.FindState(CapabilityOnline, "online")
	if cap == nil || cap.State == nil {
		return false
	}
	var v bool
	if err := json.Unmarshal(cap.State.Value, &v); err != nil {
		return false
	}
	return v
}

// PowerState returns the power state (true = on, false = off) and whether the
// capability was present in the response.
func (r *DeviceStateResponse) PowerState() (on bool, found bool) {
	cap := r.FindState(CapabilityOnOff, "powerSwitch")
	if cap == nil || cap.State == nil {
		return false, false
	}
	var v int
	if err := json.Unmarshal(cap.State.Value, &v); err != nil {
		return false, false
	}
	return v == 1, true
}

// Brightness returns the current brightness percentage and whether the capability
// was present in the response.
func (r *DeviceStateResponse) Brightness() (percent int, found bool) {
	cap := r.FindState(CapabilityRange, "brightness")
	if cap == nil || cap.State == nil {
		return 0, false
	}
	var v int
	if err := json.Unmarshal(cap.State.Value, &v); err != nil {
		return 0, false
	}
	return v, true
}

// ColorRGB returns the current color as separate R, G, B components decoded from
// the packed integer (r<<16|g<<8|b), plus whether the capability was present.
func (r *DeviceStateResponse) ColorRGB() (red, green, blue int, found bool) {
	cap := r.FindState(CapabilityColorSetting, "colorRgb")
	if cap == nil || cap.State == nil {
		return 0, 0, 0, false
	}
	var packed int
	if err := json.Unmarshal(cap.State.Value, &packed); err != nil {
		return 0, 0, 0, false
	}
	return (packed >> 16) & 0xFF, (packed >> 8) & 0xFF, packed & 0xFF, true
}

// ColorTemp returns the current color temperature in Kelvin and whether the
// capability was present in the response.
func (r *DeviceStateResponse) ColorTemp() (kelvin int, found bool) {
	cap := r.FindState(CapabilityColorSetting, "colorTemperatureK")
	if cap == nil || cap.State == nil {
		return 0, false
	}
	var v int
	if err := json.Unmarshal(cap.State.Value, &v); err != nil {
		return 0, false
	}
	return v, true
}

// deviceStatePayload is the POST body for /device/state.
type deviceStatePayload struct {
	SKU      string `json:"sku"`
	DeviceID string `json:"device"`
}

// GetDeviceState queries the current state of a device. Each capability in the
// returned DeviceStateResponse has a State field containing its current value.
func (c *Client) GetDeviceState(ctx context.Context, sku, deviceID string) (*DeviceStateResponse, error) {
	data, err := c.doPost(ctx, "/device/state", deviceStatePayload{
		SKU:      sku,
		DeviceID: deviceID,
	})
	if err != nil {
		return nil, err
	}

	var resp DeviceStateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("govee: failed to decode device state: %w", err)
	}
	return &resp, nil
}
