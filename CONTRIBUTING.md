# Contributing to govee-go

Thanks for your interest in contributing! This project is a small, focused Go library, so the process is straightforward.

## Getting Started

1. Fork the repository and clone your fork:

   ```bash
   git clone https://github.com/<your-username>/govee-go.git
   cd govee-go
   ```

2. Make sure tests pass before you start:

   ```bash
   go test ./...
   ```

3. Create a branch for your change:

   ```bash
   git checkout -b my-change
   ```

## Making Changes

### Code style

- Run `gofmt` (or `goimports`) before committing. CI will reject unformatted code.
- Follow the conventions already in the codebase: `context.Context` as the first parameter, error wrapping with `%w`, functional options for configuration.
- Keep the dependency list at zero. Everything should use the Go standard library.

### Wire-format strings

The Govee API has some non-obvious naming. In particular, the color temperature field is `"colorTem"` (not `"colorTemp"`) in all JSON payloads. When adding support for new API fields, always verify the exact string against the [Govee Developer API Reference](https://govee-public.s3.amazonaws.com/developer-docs/GoveeAPIReference.pdf).

### Tests

- All changes should include tests. The project uses `net/http/httptest` to mock the Govee API -- no real API calls are made during tests.
- Mock responses must match the actual Govee API wire format (field names, nesting, types).
- Run the full suite before opening a PR:

  ```bash
  go test -v -race ./...
  ```

### Commits

- Write clear, concise commit messages that describe _why_, not just _what_.
- Keep commits focused -- one logical change per commit.

## Submitting a Pull Request

1. Push your branch to your fork.
2. Open a pull request against `main`.
3. Describe what you changed and why. If it fixes a bug, include steps to reproduce.
4. Make sure all tests pass in CI.

## Reporting Bugs

Open an issue with:

- What you expected to happen.
- What actually happened (include error messages or unexpected responses).
- The device model and firmware version, if relevant.
- A minimal code snippet that reproduces the problem.

## Scope

This library wraps the [Govee Developer API v1](https://developer.govee.com/docs). Contributions that stay within this scope are welcome:

- Bug fixes
- Missing API field support
- Better error messages or validation
- Documentation improvements
- Test coverage improvements

For large changes (new abstractions, breaking API changes, v2 API support), please open an issue first to discuss the approach.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
