package govee

import (
	"context"
	"encoding/json"
	"fmt"
)

// Device type constants.
const (
	DeviceLight         = "devices.types.light"
	DeviceAirPurifier   = "devices.types.air_purifier"
	DeviceThermometer   = "devices.types.thermometer"
	DeviceSocket        = "devices.types.socket"
	DeviceSensor        = "devices.types.sensor"
	DeviceHeater        = "devices.types.heater"
	DeviceHumidifier    = "devices.types.humidifier"
	DeviceDehumidifier  = "devices.types.dehumidifier"
	DeviceIceMaker      = "devices.types.ice_maker"
	DeviceAromaDiffuser = "devices.types.aroma_diffuser"
	DeviceBox           = "devices.types.box"
)

// Capability type constants.
const (
	CapabilityOnOff               = "devices.capabilities.on_off"
	CapabilityToggle              = "devices.capabilities.toggle"
	CapabilityRange               = "devices.capabilities.range"
	CapabilityMode                = "devices.capabilities.mode"
	CapabilityColorSetting        = "devices.capabilities.color_setting"
	CapabilitySegmentColorSetting = "devices.capabilities.segment_color_setting"
	CapabilityMusicSetting        = "devices.capabilities.music_setting"
	CapabilityDynamicScene        = "devices.capabilities.dynamic_scene"
	CapabilityWorkMode            = "devices.capabilities.work_mode"
	CapabilityTemperatureSetting  = "devices.capabilities.temperature_setting"
	CapabilityOnline              = "devices.capabilities.online"
	CapabilityProperty            = "devices.capabilities.property"
	CapabilityEvent               = "devices.capabilities.event"
)

// Device represents a Govee device and its capabilities as returned by GetDevices.
type Device struct {
	SKU          string       `json:"sku"`
	DeviceID     string       `json:"device"`
	DeviceName   string       `json:"deviceName,omitempty"`
	Type         string       `json:"type,omitempty"`
	Capabilities []Capability `json:"capabilities"`
}

// FindCapability returns the first capability matching the given type and instance,
// or nil if the device does not have that capability.
func (d Device) FindCapability(capType, instance string) *Capability {
	for i := range d.Capabilities {
		if d.Capabilities[i].Type == capType && d.Capabilities[i].Instance == instance {
			return &d.Capabilities[i]
		}
	}
	return nil
}

// HasCapability reports whether the device advertises the given capability type and instance.
func (d Device) HasCapability(capType, instance string) bool {
	return d.FindCapability(capType, instance) != nil
}

// Capability represents a single device capability with its parameter schema and
// optional current state (populated by GetDeviceState).
type Capability struct {
	Type       string           `json:"type"`
	Instance   string           `json:"instance"`
	Parameters json.RawMessage  `json:"parameters,omitempty"`
	State      *CapabilityState `json:"state,omitempty"`
	AlarmType  int              `json:"alarmType,omitempty"`
	EventState json.RawMessage  `json:"eventState,omitempty"`
}

// CapabilityState holds the current value of a capability, as returned by GetDeviceState.
type CapabilityState struct {
	Value json.RawMessage `json:"value"`
}

// EnumOption is a single named option within an ENUM capability parameter.
type EnumOption struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

// EnumParameters describes the valid values for an ENUM-typed capability parameter.
type EnumParameters struct {
	DataType string       `json:"dataType"`
	Options  []EnumOption `json:"options"`
}

// GetDevices returns all devices associated with the API key along with their capabilities.
// Use the returned Device values to discover what each device supports and to build
// control commands.
func (c *Client) GetDevices(ctx context.Context) ([]Device, error) {
	data, err := c.doGet(ctx, "/user/devices")
	if err != nil {
		return nil, err
	}

	var devices []Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("govee: failed to decode devices: %w", err)
	}
	return devices, nil
}
