# govee-go

An unofficial Go client library for the [Govee Developer REST API](https://developer.govee.com/docs), with an emphasis on simplicity.

## Installation

```shell
go get github.com/DTCurrie/govee-go
```

## Getting an API Key

1. Open the Govee Home app on your phone.
2. Go to **Profile → Settings → Apply for API Key**.
3. Submit your information and wait for the key to arrive by email.

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    govee "github.com/DTCurrie/govee-go"
)

func main() {
    client := govee.New("your-api-key")

    devices, err := client.GetDevices(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d devices\n", len(devices))

    for _, d := range devices {
        fmt.Printf("  %s (%s) — %s\n", d.DeviceName, d.Model, d.DeviceID)
    }

    // Turn on the first device
    if len(devices) > 0 {
        d := devices[0]
        if err := client.TurnOn(context.Background(), d.DeviceID, d.Model); err != nil {
            log.Fatal(err)
        }

        // Set brightness to 75%
        client.SetBrightness(context.Background(), d.DeviceID, d.Model, 75)

        // Set color to orange
        client.SetColor(context.Background(), d.DeviceID, d.Model, 255, 165, 0)

        // Set color temperature to daylight (6500K)
        client.SetColorTemp(context.Background(), d.DeviceID, d.Model, 6500)
    }
}
```

## API Reference

### Creating a Client

```go
client := govee.New(apiKey string, opts ...Option)
```

Available options:

| Option                                  | Description                                    |
| --------------------------------------- | ---------------------------------------------- |
| `govee.WithBaseURL(url string)`         | Override the API base URL (useful for testing) |
| `govee.WithHTTPClient(hc *http.Client)` | Replace the default HTTP client                |

### Device Discovery

```go
devices, err := client.GetDevices(ctx context.Context) ([]Device, error)
```

Returns all devices associated with the API key. Each `Device` includes:

| Field                  | Type             | Description                                                                     |
| ---------------------- | ---------------- | ------------------------------------------------------------------------------- |
| `DeviceID`             | `string`         | MAC address, used in all control/state calls                                    |
| `Model`                | `string`         | Product model (e.g. `"H6159"`)                                                  |
| `DeviceName`           | `string`         | User-assigned name from the Govee app                                           |
| `Controllable`         | `bool`           | Whether the device accepts control commands                                     |
| `Retrievable`          | `bool`           | Whether the device state can be queried                                         |
| `SupportCmds`          | `[]string`       | Commands the device accepts: `"turn"`, `"brightness"`, `"color"`, `"colorTem"` |
| `Properties.ColorTemp` | `*ColorTemRange` | Color temperature range in Kelvin (nil if not supported)                        |

Use `device.SupportsCmd("color")` to check command support.

### Querying Device State

```go
state, err := client.GetDeviceState(ctx context.Context, deviceID, model string) (*DeviceState, error)
```

Only available when `device.Retrievable == true`. Returns:

| Field        | Type     | Description                                                         |
| ------------ | -------- | ------------------------------------------------------------------- |
| `Online`     | `bool`   | Whether the device is reachable (cached; may be stale)              |
| `PowerState` | `string` | `"on"` or `"off"`                                                   |
| `Brightness` | `int`    | Current brightness (0–100)                                          |
| `Color`      | `*Color` | Current RGB color (`nil` if not in color mode)                      |
| `ColorTemp`  | `int`    | Current color temperature in Kelvin (`0` if not in color temp mode) |

### Controlling Devices

All control methods require `deviceID` (MAC address) and `model`.

```go
err = client.TurnOn(ctx, deviceID, model string) error
err = client.TurnOff(ctx, deviceID, model string) error
err = client.SetBrightness(ctx, deviceID, model string, brightness int) error  // 0-100
err = client.SetColor(ctx, deviceID, model string, r, g, b int) error          // 0-255 per channel
err = client.SetColorTemp(ctx, deviceID, model string, kelvin int) error
```

### Error Handling

API errors are returned as `*APIError`:

```go
if err != nil {
    var apiErr *govee.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error %d: %s\n", apiErr.Code, apiErr.Message)
    }
}
```

## Rate Limits

| Scope             | Limit                         |
| ----------------- | ----------------------------- |
| All APIs combined | 10,000 requests/day           |
| `GetDevices`      | 10 requests/minute            |
| `GetDeviceState`  | 10 requests/minute per device |
| Control commands  | 10 requests/minute per device |

## License

MIT — see [LICENSE](LICENSE).
