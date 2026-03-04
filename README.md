# govee-go

A Go client library for the [Govee API](https://developer.govee.com/).

The Govee OpenAPI uses a **capability-based model**: each device self-describes the
controls it supports (power, brightness, color, scenes, etc.) together with valid
parameter ranges and options. All control commands are sent through a single endpoint
using a `{type, instance, value}` triple.

## Installation

```
go get github.com/DTCurrie/govee-go
```

## Getting an API Key

1. Open the Govee app on your phone.
2. Go to **Profile → About Us → Apply for API Key**.
3. Fill out the form. The key is emailed within a few minutes.

## Usage

### Creating a client

```go
import govee "github.com/DTCurrie/govee-go"

client := govee.New("your-api-key")
```

### Discovering devices

```go
devices, err := client.GetDevices(ctx)
if err != nil {
    log.Fatal(err)
}
for _, d := range devices {
    fmt.Printf("%s  sku=%s  id=%s\n", d.DeviceName, d.SKU, d.DeviceID)
    for _, cap := range d.Capabilities {
        fmt.Printf("  capability: type=%s instance=%s\n", cap.Type, cap.Instance)
    }
}
```

### Checking device capabilities

Use the helper methods to avoid iterating manually:

```go
if d.HasCapability(govee.CapabilityColorSetting, "colorRgb") {
    fmt.Println("device supports RGB color")
}
```

### Controlling devices

All control methods accept `(ctx, sku, deviceID)` as the first three arguments.

```go
sku      := "H6008"
deviceID := "AA:BB:CC:DD:EE:FF:00:00"

// Power
_ = client.TurnOn(ctx, sku, deviceID)
_ = client.TurnOff(ctx, sku, deviceID)

// Brightness (1–100)
_ = client.SetBrightness(ctx, sku, deviceID, 75)

// Color (RGB, each 0–255)
_ = client.SetColor(ctx, sku, deviceID, 255, 128, 0)

// Color temperature (Kelvin)
_ = client.SetColorTemp(ctx, sku, deviceID, 4000)

// Named toggle feature (e.g. gradient strip, nightlight)
_ = client.SetToggle(ctx, sku, deviceID, "gradientToggle", true)

// Work mode (appliances such as humidifiers, air purifiers)
_ = client.SetWorkMode(ctx, sku, deviceID, govee.WorkModeValue{WorkMode: 1, ModeValue: 2})

// Target temperature (kettles, heaters)
_ = client.SetTemperature(ctx, sku, deviceID, govee.TemperatureValue{Temperature: 100, Unit: "C"})

// Generic capability (advanced use)
_ = client.ControlDevice(ctx, sku, deviceID, govee.CapabilityCommand{
    Type:     govee.CapabilityRange,
    Instance: "brightness",
    Value:    50,
})
```

### Scenes

#### Listing and activating built-in dynamic scenes

Dynamic scenes (animated color loops, effects, etc.) are fetched separately from
the device list because a device can have hundreds of them:

```go
scenes, err := client.GetScenes(ctx, sku, deviceID)
if err != nil {
    log.Fatal(err)
}
for _, s := range scenes {
    fmt.Printf("scene: %s  id=%d paramId=%d\n", s.Name, s.Value.ID, s.Value.ParamID)
}

// Activate a scene
if len(scenes) > 0 {
    _ = client.SetLightScene(ctx, sku, deviceID, scenes[0].Value)
}
```

#### DIY scenes

```go
diyScenes, err := client.GetDIYScenes(ctx, sku, deviceID)
if err != nil {
    log.Fatal(err)
}
if len(diyScenes) > 0 {
    _ = client.SetDIYScene(ctx, sku, deviceID, diyScenes[0].Value)
}
```

#### Snapshots

```go
_ = client.SetSnapshot(ctx, sku, deviceID, 1)
```

### Music mode

```go
autoColor := 1
_ = client.SetMusicMode(ctx, sku, deviceID, govee.MusicModeValue{
    MusicMode:   3,
    Sensitivity: 80,
    AutoColor:   &autoColor,
})
```

### Segment control (light strips with individually-addressable segments)

```go
// Color segments 0, 1, and 2 red
_ = client.SetSegmentColor(ctx, sku, deviceID, []int{0, 1, 2}, 0xFF0000)

// Set segment 3 brightness to 50%
_ = client.SetSegmentBrightness(ctx, sku, deviceID, []int{3}, 50)
```

### Querying device state

```go
state, err := client.GetDeviceState(ctx, sku, deviceID)
if err != nil {
    log.Fatal(err)
}

fmt.Println("online:", state.IsOnline())

if on, ok := state.PowerState(); ok {
    fmt.Println("power:", on)
}
if pct, ok := state.Brightness(); ok {
    fmt.Println("brightness:", pct)
}
if r, g, b, ok := state.ColorRGB(); ok {
    fmt.Printf("color: rgb(%d,%d,%d)\n", r, g, b)
}
if k, ok := state.ColorTemp(); ok {
    fmt.Println("color temp:", k, "K")
}

// Low-level access to any capability state
if cap := state.FindState(govee.CapabilityWorkMode, "workMode"); cap != nil {
    fmt.Printf("work mode state: %s\n", cap.State.Value)
}
```

### Real-time device events via MQTT

`EventClient` subscribes to the Govee MQTT broker and calls your handler for each
device state update:

```go
ec := govee.NewEventClient("your-api-key", func(event govee.DeviceEvent) {
    fmt.Printf("event from %s (%s)\n", event.DeviceName, event.SKU)
    for _, cap := range event.Capabilities {
        fmt.Printf("  %s/%s = %s\n", cap.Type, cap.Instance, cap.State)
    }
})

if err := ec.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer ec.Close()

// Block until done
<-ctx.Done()
```

## Rate limits

The API allows up to **10,000 requests per account per day**. A `*govee.APIError`
with `Code == 429` is returned when the limit is exceeded.

## Error handling

All API errors are returned as `*govee.APIError`:

```go
devices, err := client.GetDevices(ctx)
if err != nil {
    var apiErr *govee.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error %d: %s\n", apiErr.Code, apiErr.Message)
    }
    log.Fatal(err)
}
```

## Capability type constants

The package exports string constants for all well-known capability types:

| Constant                        | Value                                        |
| ------------------------------- | -------------------------------------------- |
| `CapabilityOnOff`               | `devices.capabilities.on_off`                |
| `CapabilityToggle`              | `devices.capabilities.toggle`                |
| `CapabilityRange`               | `devices.capabilities.range`                 |
| `CapabilityMode`                | `devices.capabilities.mode`                  |
| `CapabilityColorSetting`        | `devices.capabilities.color_setting`         |
| `CapabilitySegmentColorSetting` | `devices.capabilities.segment_color_setting` |
| `CapabilityMusicSetting`        | `devices.capabilities.music_setting`         |
| `CapabilityDynamicScene`        | `devices.capabilities.dynamic_scene`         |
| `CapabilityWorkMode`            | `devices.capabilities.work_mode`             |
| `CapabilityTemperatureSetting`  | `devices.capabilities.temperature_setting`   |
| `CapabilityOnline`              | `devices.capabilities.online`                |
| `CapabilityProperty`            | `devices.capabilities.property`              |
| `CapabilityEvent`               | `devices.capabilities.event`                 |
