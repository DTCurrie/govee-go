package govee_test

import (
	"context"
	"errors"
	"fmt"
	"log"

	govee "github.com/DTCurrie/govee-go"
)

func ExampleNew() {
	client := govee.New("your-api-key")
	_ = client
}

func ExampleNew_withOptions() {
	client := govee.New(
		"your-api-key",
		govee.WithBaseURL("https://developer-api.govee.com"),
	)
	_ = client
}

func ExampleClient_GetDevices() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range devices {
		fmt.Printf("%s (%s) — %s\n", d.DeviceName, d.Model, d.DeviceID)
	}
}

func ExampleClient_GetDeviceState() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range devices {
		if !d.Retrievable {
			continue
		}
		state, err := client.GetDeviceState(context.Background(), d.DeviceID, d.Model)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: online=%v power=%s brightness=%d\n",
			d.DeviceName, state.Online, state.PowerState, state.Brightness)
	}
}

func ExampleClient_TurnOn() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		return
	}

	d := devices[0]
	if err := client.TurnOn(context.Background(), d.DeviceID, d.Model); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_TurnOff() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		return
	}

	d := devices[0]
	if err := client.TurnOff(context.Background(), d.DeviceID, d.Model); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_SetBrightness() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		return
	}

	d := devices[0]
	if err := client.SetBrightness(context.Background(), d.DeviceID, d.Model, 75); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_SetColor() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		return
	}

	d := devices[0]
	// Set color to orange (R=255, G=165, B=0).
	if err := client.SetColor(context.Background(), d.DeviceID, d.Model, 255, 165, 0); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_SetColorTemp() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		return
	}

	d := devices[0]
	// Check that the device supports color temperature before setting it.
	if !d.SupportsCmd("colorTem") {
		return
	}
	// Set to daylight (6500K). The valid range is in d.Properties.ColorTemp.
	if err := client.SetColorTemp(context.Background(), d.DeviceID, d.Model, 6500); err != nil {
		log.Fatal(err)
	}
}

func ExampleDevice_SupportsCmd() {
	client := govee.New("your-api-key")

	devices, err := client.GetDevices(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range devices {
		if d.SupportsCmd("color") {
			fmt.Printf("%s supports RGB color\n", d.DeviceName)
		}
	}
}

func ExampleAPIError() {
	client := govee.New("invalid-key")

	_, err := client.GetDevices(context.Background())
	if err != nil {
		var apiErr *govee.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("API error %d: %s\n", apiErr.Code, apiErr.Message)
		} else {
			log.Fatal(err)
		}
	}
}
