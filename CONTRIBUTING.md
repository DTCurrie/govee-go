# Contributing

This project uses [Task](https://taskfile.dev) as its task runner. See the [installation docs](https://taskfile.dev/docs/installation) for instructions.

## Setup

```
git clone https://github.com/DTCurrie/govee-go.git
cd govee-go
task setup
```

## Common tasks

| Command        | What it does                                             |
| -------------- | -------------------------------------------------------- |
| `task test`    | Run the full test suite with the race detector           |
| `task lint`    | Run golangci-lint (pinned version from go.mod)           |
| `task fmt`     | Check formatting                                         |
| `task fmt:fix` | Reformat all Go source files with gofmt                  |
| `task vet`     | Run go vet static analysis                               |
| `task build`   | Compile the package                                      |
| `task docs`    | Generate the API reference into `./www`                  |
| `task check`   | Run fmt, vet, lint, and test — the full pre-commit check |

Run `task check` before submitting a pull request.

## Project conventions

### Package structure

| File         | Purpose                                                               |
| ------------ | --------------------------------------------------------------------- |
| `govee.go`   | `Client` struct, HTTP helpers (`doGet`, `doPost`), functional options |
| `devices.go` | `Device`, `Capability`, and related types; `GetDevices`               |
| `control.go` | `ControlDevice` + all typed control helpers                           |
| `state.go`   | `GetDeviceState` + `DeviceStateResponse` convenience methods          |
| `scenes.go`  | `GetScenes`, `GetDIYScenes`                                           |
| `events.go`  | `EventClient` for MQTT subscriptions                                  |
| `errors.go`  | `APIError` type                                                       |

### Device identification

Devices are always identified by the `(sku, deviceID)` pair. These correspond to
the `sku` and `device` fields in the API. Both values come from `GetDevices`.

### Capability model

The Govee API uses a capability-based model. Control commands are always a
`{type, instance, value}` triple matching what a device advertises in its
`Capabilities` slice. The `ControlDevice` method is the generic form; prefer the
typed helpers (`TurnOn`, `SetBrightness`, etc.) for well-known capabilities.

### Adding new capability helpers

1. Add a helper method to `control.go` that calls `ControlDevice` with the
   correct `type`/`instance`/`value`.
2. If the value is a struct (not a scalar), define a named value type in
   `control.go` (e.g. `MusicModeValue`, `WorkModeValue`).
3. Add a test in `control_test.go` that verifies the correct capability fields are
   sent.

### HTTP transport

`doGet` handles `GET /user/devices` and uses the `{code, message, data}` envelope.
`doPost` handles all other endpoints and uses the `{requestId, payload}` request
envelope and `{requestId, code, msg/message, payload}` response envelope.

Non-200 `code` values inside the response envelope are returned as `*APIError` even
when the HTTP status is 200. This matches Govee's API behavior.

### Testing

- Use `net/http/httptest` for HTTP mocking — no external test dependencies.
- Keep tests in the `govee_test` external package to test the public API.
- Shared test helpers live in `testutils_test.go`.
- Use table-driven tests where multiple input/output combinations are tested.
- Call `t.Parallel()` on independent tests and sub-tests.
- Use `t.Helper()` in test helper functions.
- Use `t.Cleanup()` for teardown (e.g., closing test servers).

### Dependencies

Runtime dependencies are intentionally minimal. The only non-standard dependency is
`github.com/eclipse/paho.mqtt.golang` for MQTT event subscriptions. Avoid adding
new runtime dependencies without a strong reason.

`golangci-lint` and `doc2go` are tool-only dependencies declared with Go 1.24+
`tool` directives in `go.mod`.
