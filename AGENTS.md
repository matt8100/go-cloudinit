# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go library for generating and validating `#cloud-config` YAML. Source is organized by cloud-init domain:

- `config.go`, `render.go`, `validate.go`, `validate_modules.go`: top-level API, rendering, and validation.
- `commands.go`, `files.go`, `users.go`, `ssh.go`, `certs.go`, `packages.go`, `storage.go`, `system.go`, `integrations.go`: typed module definitions grouped by feature area.
- `cloudinit_test.go`: unit and regression tests.
- `integration_test.go`: optional schema verification against `cloud-init`.
- `README.md`: public usage example and project scope.

## Build, Test, and Development Commands
- `gofmt -w *.go`: format all Go files in the repo.
- `go test ./...`: run the full test suite; this is the main development check.
- `go test ./... -run TestName`: run a focused test while iterating.
- `cloud-init schema -c <file> -t cloud-config`: optional external schema check when `cloud-init` is installed.

This repo does not ship a binary. Treat `go test ./...` as the compile-and-verify command.

## Coding Style & Naming Conventions
Use standard Go formatting with tabs via `gofmt`; do not hand-format indentation. Keep the public API typed and close to cloud-init semantics: exported Go names use `CamelCase`, YAML tags mirror upstream keys exactly, and validation errors should include precise field paths such as `apt.sources.example`.

Prefer small, domain-focused files over monolithic ones. When fixing a bug, add the narrowest possible validation or rendering change and keep runtime behavior pure Go.

## Testing Guidelines
Tests use Go’s standard `testing` package. Name tests as `TestXxx`, and add regression tests for every schema or validation bug you fix. Put most cases in `cloudinit_test.go`; use `integration_test.go` only for checks that depend on the external `cloud-init` binary.

## Commit & Pull Request Guidelines
There is no established Git history yet. Use short, imperative commit subjects such as `validate apt source keys`. Pull requests should explain user-visible behavior changes, list validation/rendering impacts, and note the verification run (for example, `gofmt -w *.go` and `go test ./...`). Screenshots are not relevant for this library.
