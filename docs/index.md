---
title: "Bambulabs API"
---

# Bambulabs API

Welcome to the Bambulabs API documentation. This repository provides a Go client library and related internal tooling to interact with Bambulabs printers via the vendor MQTT protocol and local emulation tools. The docs are written as plain Markdown so they can be used with Hugo or another static site generator (SSG).

Contents

- Overview and architecture
- Quickstart guide (get started in minutes)
- API reference (detailed types, methods, and examples)
- Supported models and capabilities
- Deployment: Publish this site with Hugo or any SSG

Project layout

- `bambulabs.go`, `printer.go`, `lights.go`, `fans.go`, `state.go` — core client API
- `internal/mqtt` — MQTT client and message handling
- `internal/ftp` — FTP client and file operations
- `internal/protocol` — command and payload helpers
- `hms` — hardware model/service helpers and generators
- `internal/emulator` — local emulator for development & testing
- `docs/` — this site content

Goals

- Provide a stable, idiomatic Go client for interacting with Bambulabs printers
- Make it easy to script printer control (lights, fans, G-code) and read state
- Provide a developer-friendly emulator and test harness for CI and local development

Where to start

- Read the Quickstart: [Quickstart](quickstart.md)
- Browse the API reference: [API Reference](api.md)
- See supported models: [Supported Models](supported_models.md)

Contributing

See CONTRIBUTING.md for guidelines on reporting issues, submitting changes, and test expectations.

License

This project is licensed under the terms in LICENSE.
# Bambulabs API documentation index

- [Quickstart](quickstart.md)
