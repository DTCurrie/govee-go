package govee

import (
	"context"
	"fmt"
)

// controlRequest is the body sent to PUT /v1/devices/control.
type controlRequest struct {
	Device string     `json:"device"`
	Model  string     `json:"model"`
	Cmd    controlCmd `json:"cmd"`
}

// controlCmd is the cmd field in a control request.
type controlCmd struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// sendCommand sends a control command to a device.
func (c *Client) sendCommand(ctx context.Context, deviceID, model, name string, value any) error {
	_, err := c.doPut(ctx, "/v1/devices/control", controlRequest{
		Device: deviceID,
		Model:  model,
		Cmd: controlCmd{
			Name:  name,
			Value: value,
		},
	})
	return err
}

// TurnOn turns the device on.
func (c *Client) TurnOn(ctx context.Context, deviceID, model string) error {
	return c.sendCommand(ctx, deviceID, model, "turn", "on")
}

// TurnOff turns the device off.
func (c *Client) TurnOff(ctx context.Context, deviceID, model string) error {
	return c.sendCommand(ctx, deviceID, model, "turn", "off")
}

// SetBrightness sets the brightness of the device (0–100).
// A value of 0 turns the device off.
func (c *Client) SetBrightness(ctx context.Context, deviceID, model string, brightness int) error {
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("govee: brightness %d out of range [0, 100]", brightness)
	}
	return c.sendCommand(ctx, deviceID, model, "brightness", brightness)
}

// SetColor sets the RGB color of the device. Each channel is 0–255.
func (c *Client) SetColor(ctx context.Context, deviceID, model string, r, g, b int) error {
	for _, ch := range []struct {
		name string
		val  int
	}{{"r", r}, {"g", g}, {"b", b}} {
		if ch.val < 0 || ch.val > 255 {
			return fmt.Errorf("govee: color channel %s value %d out of range [0, 255]", ch.name, ch.val)
		}
	}
	return c.sendCommand(ctx, deviceID, model, "color", Color{R: r, G: g, B: b})
}

// SetColorTemp sets the color temperature of the device in Kelvin.
// The valid range is device-specific (reported in DeviceProperties.ColorTemp).
func (c *Client) SetColorTemp(ctx context.Context, deviceID, model string, kelvin int) error {
	if kelvin <= 0 {
		return fmt.Errorf("govee: color temperature %d must be a positive value in Kelvin", kelvin)
	}
	return c.sendCommand(ctx, deviceID, model, "colorTem", kelvin)
}
