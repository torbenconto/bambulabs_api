---
title: "Supported Models & Capabilities"
---

# Supported Models & Capabilities

This document describes how printer models and capabilities are represented and how to extend support for new models.

Model representation

The codebase defines a `Model` type (see `bambulabs.go`) that enumerates known printer models. Model values are used to gate features such as lights, fans, or specialized commands.

Capabilities

Capabilities are implemented as small helper functions (see `fans.go`, `lights.go`) which evaluate whether a given `Model` supports a specific feature. Typical capability checks include:

- `SupportsLight(model Model, light Light) bool`
- `SupportsFan(model Model, fan Fan) bool`

Extending support for new models

1. Add the model to the `Model` enum in `bambulabs.go`.
2. Update capability helper functions in `lights.go` and `fans.go` to include the new model where applicable.
3. Add any model-specific defaults (e.g., default number of fans or LED nodes) in `hms` or `state.go` as required.

Feature matrix (example)

| Feature | Model A | Model B | Model C |
|---|---:|---:|---:|
| Front LED | ✓ | ✓ | ✗ |
| Rear Fan | ✗ | ✓ | ✓ |
| Dual Extruder | ✗ | ✗ | ✓ |

Notes

- Where features are unknown for a model, the library errs on the side of safety: calls that would affect unsupported features return `ErrLightNotSupported` or `ErrFanNotSupported`.
- If you add model-specific behaviors beyond capability checks (e.g., different command sequences), keep them isolated and documented.

# All printer models supported

- ModelUnknown
- ModelA1Mini
- ModelA1
- ModelA2L
- ModelP1S
- ModelP2S
- ModelX1E
- ModelX1C
- ModelH2S
- ModelH2D
- ModelH2DPro
- ModelH2
- ModelH2C
- ModelX2D
