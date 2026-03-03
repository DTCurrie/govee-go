package govee

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

// Device represents a Govee smart device returned by the API.
type Device struct {
	// DeviceID is the MAC address of the device, used to identify it in control and state requests.
	DeviceID string `json:"device"`
	// Model is the product model string (e.g. "H6159").
	Model string `json:"model"`
	// DeviceName is the user-assigned name from the Govee app.
	DeviceName string `json:"deviceName"`
	// Controllable is true when the device accepts control commands.
	Controllable bool `json:"controllable"`
	// Retrievable is true when the device state can be queried.
	Retrievable bool `json:"retrievable"`
	// SupportCmds lists the commands the device accepts: "turn", "brightness", "color", "colorTem".
	SupportCmds []string `json:"supportCmds"`
	// Properties holds optional per-device property constraints.
	Properties DeviceProperties `json:"properties"`
}

// SupportsCmd reports whether the device supports the given command name.
func (d Device) SupportsCmd(cmd string) bool {
	return slices.Contains(d.SupportCmds, cmd)
}

// DeviceProperties holds optional per-device constraints returned by DeviceList.
type DeviceProperties struct {
	// ColorTemp is non-nil when the device supports the "colorTem" command.
	ColorTemp *ColorTempRange `json:"colorTemp,omitempty"`
}

// ColorTempRange is the supported color temperature range for a device, in Kelvin.
type ColorTempRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// DeviceState is the current reported state of a device.
type DeviceState struct {
	// Online indicates whether the device is reachable. Note: this is cached by
	// the Govee API and may be stale; control commands should still be attempted
	// even when Online is false.
	Online bool
	// PowerState is "on" or "off".
	PowerState string
	// Brightness is the current brightness level (0–100). 0 means off or unknown.
	Brightness int
	// Color is the current RGB color, or nil if not reported.
	Color *Color
	// ColorTemp is the current color temperature in Kelvin, or 0 if not reported.
	ColorTemp int
}

// Color represents an RGB color value.
type Color struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// deviceListResponse is the shape of the data field from GET /v1/devices.
type deviceListResponse struct {
	Devices []deviceJSON `json:"devices"`
}

// deviceJSON mirrors the raw API device object for unmarshalling. Properties
// uses a custom struct to accommodate the nested colorTemp range.
type deviceJSON struct {
	Device       string          `json:"device"`
	Model        string          `json:"model"`
	DeviceName   string          `json:"deviceName"`
	Controllable bool            `json:"controllable"`
	Retrievable  bool            `json:"retrievable"`
	SupportCmds  []string        `json:"supportCmds"`
	Properties   devicePropsJSON `json:"properties"`
}

type devicePropsJSON struct {
	ColorTemp *colorTemRangeJSON `json:"colorTem,omitempty"`
}

type colorTemRangeJSON struct {
	Range struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"range"`
}

// GetDevices returns all Govee devices associated with the API key.
func (c *Client) GetDevices(ctx context.Context) ([]Device, error) {
	data, err := c.doGet(ctx, "/v1/devices", nil)
	if err != nil {
		return nil, err
	}

	var resp deviceListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("govee: failed to decode device list: %w", err)
	}

	devices := make([]Device, 0, len(resp.Devices))
	for _, d := range resp.Devices {
		dev := Device{
			DeviceID:     d.Device,
			Model:        d.Model,
			DeviceName:   d.DeviceName,
			Controllable: d.Controllable,
			Retrievable:  d.Retrievable,
			SupportCmds:  d.SupportCmds,
		}
		if d.Properties.ColorTemp != nil {
			dev.Properties.ColorTemp = &ColorTempRange{
				Min: d.Properties.ColorTemp.Range.Min,
				Max: d.Properties.ColorTemp.Range.Max,
			}
		}
		devices = append(devices, dev)
	}
	return devices, nil
}
