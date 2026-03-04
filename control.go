package govee

import (
	"context"
	"fmt"
)

// CapabilityCommand is a capability control command sent to a device via ControlDevice.
type CapabilityCommand struct {
	Type     string      `json:"type"`
	Instance string      `json:"instance"`
	Value    interface{} `json:"value"`
}

// controlRequest is the POST body for /device/control.
type controlRequest struct {
	SKU        string            `json:"sku"`
	DeviceID   string            `json:"device"`
	Capability CapabilityCommand `json:"capability"`
}

// LightSceneValue identifies a dynamic built-in light scene by its API-assigned ID and
// parameter ID. Obtain these values from GetScenes and pass to SetLightScene.
type LightSceneValue struct {
	ID      int `json:"id"`
	ParamID int `json:"paramId"`
}

// MusicModeValue configures a music visualization mode.
type MusicModeValue struct {
	MusicMode   int  `json:"musicMode"`
	Sensitivity int  `json:"sensitivity"`
	AutoColor   *int `json:"autoColor,omitempty"`
	RGB         *int `json:"rgb,omitempty"`
}

// WorkModeValue sets the active working mode and its associated sub-value for appliances.
type WorkModeValue struct {
	WorkMode  int `json:"workMode"`
	ModeValue int `json:"modeValue"`
}

// TemperatureValue sets the target temperature for heaters, thermostats, and similar devices.
type TemperatureValue struct {
	Temperature int    `json:"temperature"`
	Unit        string `json:"unit,omitempty"`
	AutoStop    *int   `json:"autoStop,omitempty"`
}

// segmentColorValue is the wire format for the segmentedColorRgb instance.
type segmentColorValue struct {
	Segment []int `json:"segment"`
	RGB     int   `json:"rgb"`
}

// segmentBrightnessValue is the wire format for the segmentedBrightness instance.
type segmentBrightnessValue struct {
	Segment    []int `json:"segment"`
	Brightness int   `json:"brightness"`
}

// ControlDevice sends a raw capability command to a device identified by (sku, deviceID).
// This is the generic form; prefer the typed helpers for common operations.
func (c *Client) ControlDevice(ctx context.Context, sku, deviceID string, cmd CapabilityCommand) error {
	_, err := c.doPost(ctx, "/device/control", controlRequest{
		SKU:        sku,
		DeviceID:   deviceID,
		Capability: cmd,
	})
	return err
}

// TurnOn powers the device on.
func (c *Client) TurnOn(ctx context.Context, sku, deviceID string) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityOnOff,
		Instance: "powerSwitch",
		Value:    1,
	})
}

// TurnOff powers the device off.
func (c *Client) TurnOff(ctx context.Context, sku, deviceID string) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityOnOff,
		Instance: "powerSwitch",
		Value:    0,
	})
}

// SetBrightness sets the device brightness. percent must be in the range [1, 100].
func (c *Client) SetBrightness(ctx context.Context, sku, deviceID string, percent int) error {
	if percent < 1 || percent > 100 {
		return fmt.Errorf("govee: brightness %d out of range [1, 100]", percent)
	}
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityRange,
		Instance: "brightness",
		Value:    percent,
	})
}

// SetColor sets the device color. Each of r, g, b must be in the range [0, 255].
func (c *Client) SetColor(ctx context.Context, sku, deviceID string, r, g, b int) error {
	if r < 0 || r > 255 {
		return fmt.Errorf("govee: red value %d out of range [0, 255]", r)
	}
	if g < 0 || g > 255 {
		return fmt.Errorf("govee: green value %d out of range [0, 255]", g)
	}
	if b < 0 || b > 255 {
		return fmt.Errorf("govee: blue value %d out of range [0, 255]", b)
	}
	rgb := (r << 16) | (g << 8) | b
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityColorSetting,
		Instance: "colorRgb",
		Value:    rgb,
	})
}

// SetColorTemp sets the device color temperature in Kelvin. kelvin must be greater than 0.
func (c *Client) SetColorTemp(ctx context.Context, sku, deviceID string, kelvin int) error {
	if kelvin <= 0 {
		return fmt.Errorf("govee: color temperature %d must be greater than 0", kelvin)
	}
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityColorSetting,
		Instance: "colorTemperatureK",
		Value:    kelvin,
	})
}

// SetToggle enables or disables a named toggle feature such as "gradientToggle" or
// "nightlightToggle".
func (c *Client) SetToggle(ctx context.Context, sku, deviceID, instance string, on bool) error {
	val := 0
	if on {
		val = 1
	}
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityToggle,
		Instance: instance,
		Value:    val,
	})
}

// SetLightScene activates a built-in light scene. value should be a LightSceneValue
// obtained from GetScenes for devices with compound scene values, or a plain int
// for devices whose scene options use simple integer IDs.
func (c *Client) SetLightScene(ctx context.Context, sku, deviceID string, value interface{}) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityDynamicScene,
		Instance: "lightScene",
		Value:    value,
	})
}

// SetDIYScene activates a user-created DIY scene. value is the scene ID obtained from
// GetDIYScenes.
func (c *Client) SetDIYScene(ctx context.Context, sku, deviceID string, value int) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityDynamicScene,
		Instance: "diyScene",
		Value:    value,
	})
}

// SetSnapshot activates a saved snapshot by its integer ID.
func (c *Client) SetSnapshot(ctx context.Context, sku, deviceID string, value int) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityDynamicScene,
		Instance: "snapshot",
		Value:    value,
	})
}

// SetMusicMode activates a music visualization mode.
func (c *Client) SetMusicMode(ctx context.Context, sku, deviceID string, mode MusicModeValue) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityMusicSetting,
		Instance: "musicMode",
		Value:    mode,
	})
}

// SetWorkMode sets the working mode of an appliance such as a humidifier or air purifier.
func (c *Client) SetWorkMode(ctx context.Context, sku, deviceID string, mode WorkModeValue) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityWorkMode,
		Instance: "workMode",
		Value:    mode,
	})
}

// SetSegmentColor sets the color of specific light strip segments.
// segments is a slice of zero-based segment indices. rgb is a packed color integer (r<<16|g<<8|b).
func (c *Client) SetSegmentColor(ctx context.Context, sku, deviceID string, segments []int, rgb int) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilitySegmentColorSetting,
		Instance: "segmentedColorRgb",
		Value:    segmentColorValue{Segment: segments, RGB: rgb},
	})
}

// SetSegmentBrightness sets the brightness of specific light strip segments.
// segments is a slice of zero-based segment indices. brightness must be in [0, 100].
func (c *Client) SetSegmentBrightness(ctx context.Context, sku, deviceID string, segments []int, brightness int) error {
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("govee: brightness %d out of range [0, 100]", brightness)
	}
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilitySegmentColorSetting,
		Instance: "segmentedBrightness",
		Value:    segmentBrightnessValue{Segment: segments, Brightness: brightness},
	})
}

// SetTemperature sets the target temperature of a heater, thermostat, or kettle.
func (c *Client) SetTemperature(ctx context.Context, sku, deviceID string, temp TemperatureValue) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityTemperatureSetting,
		Instance: "targetTemperature",
		Value:    temp,
	})
}

// SetMode sets a mode-type capability. instance identifies which mode to set
// (e.g. "nightlightScene", "presetScene"). value is the mode ID.
func (c *Client) SetMode(ctx context.Context, sku, deviceID, instance string, value int) error {
	return c.ControlDevice(ctx, sku, deviceID, CapabilityCommand{
		Type:     CapabilityMode,
		Instance: instance,
		Value:    value,
	})
}
