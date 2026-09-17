# Migrating from v0.2.1 to v0.3.0

v0.3.0 replaces raw MQTT state with typed print, fan, light, AMS, and HMS systems. File operations move to `Files()`. Fan inputs change from PWM values to percentages.

This guide targets the current v0.3.0 working tree, including the restored file client, G-code method, and `RequestUpdate` interface method. The baseline is tagged v0.2.1; later `master` changes are identified where they differ.

The module path remains `github.com/torbenconto/bambulabs_api`, and Go 1.26 remains required. Once the release tag is available:

```sh
go get github.com/torbenconto/bambulabs_api@v0.3.0
go mod tidy
go test ./...
```

Examples use `bambu "github.com/torbenconto/bambulabs_api"`. Add the standard-library imports used by each function. Legacy HMS examples also import `github.com/torbenconto/bambulabs_api/hms`.

## API changes

| v0.2.1 / legacy master | v0.3.0 |
| --- | --- |
| `client.Add(cfg)`, `NewPrinter(ctx, cfg)` | Pass `&cfg`. |
| `p.State()` | Use `Print()`, `Fans()`, `Lights()`, `AMS()`, and `HMS()`. |
| `p.SetFan(ctx, id, pwm)` | `p.Fans().Set(ctx, id, percent)`. |
| `p.SetLight(ctx, id, mode)` | `p.Lights().Set(ctx, id, mode)`. |
| `LightOn`, `LightOff`, `LightFlashing` | `LightModeOn`, `LightModeOff`, `LightModeFlashing`. |
| `AuxiliaryFan` | `AuxillaryFan`, spelled exactly as currently exported. |
| `SupportsFan`, `SupportsLight` | Check the system's `Get(id)` result. |
| `ErrFanNotSupported`, `ErrLightNotSupported` | `ErrFanUnavailable`, `ErrLightUnavailable`. |
| `p.ListFiles(path)` | `p.Files().List(path)`. |
| `p.DownloadFile(path, w)` | `p.Files().Download(path, w)`. |
| `p.UploadFile(path, r)` | `p.Files().Upload(path, r)`. |
| `p.DeleteFile(path)` | `p.Files().Delete(path)`. |
| `hms.Error` | `bambu.HMSError`. |
| `p.RequestUpdate(ctx)` | Unchanged, including through `Printer`. |
| `p.SendGcode(ctx, lines)` | Unchanged, including through `Printer`. |

## Connections

**v0.2.1:**

<!-- example: legacy -->
```go
func addPrinter(client *bambu.Client, cfg bambu.Config) (bambu.Printer, error) {
	return client.Add(cfg)
}
```

**v0.3.0:**

<!-- example: current -->
```go
func addPrinter(client *bambu.Client, cfg bambu.Config) (bambu.Printer, error) {
	return client.Add(&cfg)
}
```

Direct construction changes to `bambu.NewPrinter(ctx, &cfg)`. `NewClient(ctx)` still returns one `*Client`. `Load`, `Remove`, `Range`, `Serial`, and `Close` are unchanged.

Construction now requests a full update and waits for an initial print report. Optional devices can still be unavailable if they have not been reported.

The constructor context controls the connection's lifetime. Keep it active while using the printer, and close the owning client or directly constructed printer when finished. The ten-second MQTT operation timeout does not bound the entire initial-state wait.

## New data model

| Accessor | Public data and operations |
| --- | --- |
| `Print()` | `Info() PrintInfo`; reported state, job metadata, progress, layers, remaining time, and failure reason. |
| `Fans()` | `Get(Fan) (FanInfo, error)`; `Set(ctx, Fan, percent)`. `FanInfo` contains `Fan` and `Percent`. |
| `Lights()` | `Get(Light) (LightInfo, error)`; `Set(ctx, Light, mode)`. `LightInfo` contains `Light` and `Mode`. |
| `AMS()` | `Units() []AMS`, `Get(unitID) *AMS`, `ExternalTray() Tray`. |
| `HMS()` | `Errors() []HMSError`. Each error exposes `Attribute`, `Code`, `GetCode()`, and `Error()`. |
| `Files()` | `FileClient` with `List`, `Download`, `Upload`, and `Delete`. Can be nil. |

AMS data is decoded into `AMS`, `Tray`, `FilamentInfo`, `RFIDInfo`, and `TemperatureRequirements`. Numbers no longer require application parsing; colors use `color.RGBA`.

Reads across systems are not an atomic printer snapshot. Print/fan/light getters return values, and HMS returns an independent slice. AMS getters currently share backing data; treat their results as read-only.

## Print status

Replace raw print telemetry with `p.Print().Info()`.

**v0.2.1:**

<!-- example: legacy -->
```go
func printStatus(p bambu.Printer) {
    if state, ok := p.State(); ok {
        fmt.Println(state.Print.GcodeState, state.Print.McPercent,
            state.Print.LayerNum, state.Print.TotalLayerNum)
    }
}
```

**v0.3.0:**

<!-- example: current -->
```go
func printStatus(p bambu.Printer) {
	info := p.Print().Info()
	fmt.Println(info.State, info.ProgressPercent,
		info.CurrentLayer, info.TotalLayers, info.Remaining)
}
```

`PrintInfo` also exposes `Filename`, `Name`, `TaskID`, `SubtaskID`, and `FailureReason`. `Remaining` is a `time.Duration`, converted from the reported minute count.

`Info()` returns one snapshot value. Each print report replaces it directly using the fixture types. There are no availability flags, partial-field merging, or job-boundary heuristics.

Replace `GcodeState` with `PrintState` and constants such as `RUNNING` with `PrintStateRunning`. See [print status](print.md) for all fields and states.

## Fans

**v0.2.1:**

<!-- example: legacy -->
```go
func setAuxFan(ctx context.Context, p bambu.Printer) error {
	return p.SetFan(ctx, bambu.AuxiliaryFan, 128)
}
```

**v0.3.0:**

<!-- example: current -->
```go
func setAuxFan(ctx context.Context, p bambu.Printer) error {
	return p.Fans().Set(ctx, bambu.AuxillaryFan, 50)
}

func coolingPercent(p bambu.Printer) (int, error) {
	fan, err := p.Fans().Get(bambu.PartCoolingFan)
	return fan.Percent, err
}
```

Inputs are integers from 0 to 100, rounded to the nearest ten: `54 → 50`, `55 → 60`. Values outside that range return `ErrInvalidFanPercent`. An old PWM value of `255` becomes `100`; passing `255` unchanged now fails.

For callers that still supply PWM:

<!-- example: current -->
```go
func setFanPWM(ctx context.Context, p bambu.Printer, id bambu.Fan, pwm uint8) error {
	percent := int(math.Round(float64(pwm) * 100 / 255))
	return p.Fans().Set(ctx, id, percent)
}
```

`FanInfo.Percent` is a percentage, not PWM or the raw 0 to 15 report value. Conversion is approximate; a later printer report may differ from the requested percentage.

`Fan` changes from `uint8` to `int`, and `Fan.String()` is removed. `SecondaryAuxillaryFan` is declared but is not populated by the current decoder.

## Lights and availability

**v0.2.1:**

<!-- example: legacy -->
```go
func chamberLight(ctx context.Context, p bambu.Printer) error {
	return p.SetLight(ctx, bambu.ChamberLight, bambu.LightOn)
}
```

**v0.3.0:**

<!-- example: current -->
```go
func chamberLight(ctx context.Context, p bambu.Printer) error {
	err := p.Lights().Set(ctx, bambu.ChamberLight, bambu.LightModeOn)
	if errors.Is(err, bambu.ErrLightUnavailable) {
		return nil // This application skips lights that have not been reported.
	}
	return err
}
```

Read the mode through `p.Lights().Get(id)`. `ChamberLight2` is new; `ChamberLight` and `WorkLight` retain their names. Invalid modes return `ErrInvalidLightMode`.

Availability now follows received telemetry instead of static model tables. Both `Get` and `Set` return the relevant unavailable error for devices not yet reported. `ModelUnknown` no longer blocks all fan/light commands, although the model still affects AMS decoding defaults.

Post-tag `master` added `SetLightFlashing(ctx, light, cfg)`. That method has no replacement accepting custom timing. Use `Lights().Set(ctx, light, bambu.LightModeFlashing)` for default timing: 500 ms on, 500 ms off, one loop, 1 s interval.

## Optimistic updates

Fan/light setters update local state after validation and gate acquisition, before sending completes.

- Getters immediately expose the requested value unless a report supersedes it.
- Failed sends restore the previous value unless that device received a newer state update.
- Explicit reports override optimistic values. Omitted fan/light fields preserve state.
- Commands serialize per system. Waiting calls can be canceled.
- Setters use a ten-second deadline when none is supplied, including gate wait time.

Remove delays used only to reflect a command in the UI. A matching getter value does not confirm printer execution. The API has no separate confirmed-state getter or pending flag.

## AMS

**v0.2.1:**

<!-- example: legacy -->
```go
func printAMS(p bambu.Printer) {
	if state, ok := p.State(); ok {
		for _, unit := range state.Print.Ams.Ams {
			for _, tray := range unit.Tray {
				fmt.Println(unit.ID, tray.ID, tray.TrayType, tray.Remain)
			}
		}
	}
}
```

**v0.3.0:**

<!-- example: current -->
```go
func printAMS(p bambu.Printer) {
	for _, unit := range p.AMS().Units() {
		for _, tray := range unit.Trays {
			fmt.Println(unit.ID, tray.Slot,
				tray.Filament.Material, tray.Filament.RemainingPercent)
		}
	}
}
```

| Legacy field | New field |
| --- | --- |
| Unit `ID`, `Humidity` strings | `AMS.ID`, `AMS.HumidityLevel` integers |
| Unit `Tray` | `AMS.Trays` |
| Tray `ID` string | `Tray.Slot` integer |
| `TrayType`, `Remain`, `TrayDiameter` | `Filament.Material`, `Filament.RemainingPercent`, `Filament.Diameter` |
| `TrayColor`, `Cols` | `Filament.Color`, `Filament.Colors` |
| `TagUID`, `TrayUUID` | `RFID.UID`, `RFID.UUID` |
| `NozzleTempMin`, `NozzleTempMax`, `BedTemp` | `TemperatureInfo.MinNozzleTemp`, `TemperatureInfo.MaxNozzleTemp`, `TemperatureInfo.BedTemp` |

Replace `state.Print.VtTray` with `p.AMS().ExternalTray()`.

`AMS().Get(id)` uses the reported unit ID and returns nil if absent. IDs can be sparse, such as `128`. `AMS.Tray(slot)` currently uses a slice index; iterate over `Trays` and compare `Tray.Slot` when looking up a reported slot ID.

`HumidityLevel` is a level, not necessarily relative humidity percent. `HasFilament()` checks for positive remaining percentage or a nonempty material string. An explicit empty AMS report currently does not clear cached units.

## HMS

**v0.2.1:**

<!-- example: legacy -->
```go
func printHMS(p bambu.Printer) {
	if state, ok := p.State(); ok {
		for _, issue := range state.Print.HmsErrors {
			fmt.Println(issue.GetCode(), issue.Error())
		}
	}
}
```

**v0.3.0:**

<!-- example: current -->
```go
func printHMS(p bambu.Printer) {
	for _, issue := range p.HMS().Errors() {
		fmt.Println(issue.GetCode(), issue.Error())
	}
}
```

A received HMS list replaces the previous list. An empty list clears it; omitted or null HMS fields preserve it. `GetCode()` keeps the underscore-separated format. `Error()` resolves known descriptions and returns the code for unknown entries.

Replace the public `.../hms` import with the root package. For manual construction, the old signatures differ:

**Tagged v0.2.1:**

<!-- example: tag -->
```go
func savedHMS() *hms.Error {
	return hms.NewError("HMS_0300_0100_0001_0006")
}
```

**Post-tag master:**

<!-- example: master -->
```go
func savedHMS() *hms.Error {
	return hms.NewError(0x00010006, 0x03000100)
}
```

**v0.3.0:**

<!-- example: current -->
```go
func savedHMS() *bambu.HMSError {
	return &bambu.HMSError{Code: 0x00010006, Attribute: 0x03000100}
}
```

The implementation is now under `internal/hms`. No public string parser or replacement for its old `HmsErrors` map, `Module`, or `Severity` constants is exposed.

## Files

Move file calls to `p.Files()` using the names in the API table. `List` still returns `[]os.FileInfo`; download/upload still accept `io.Writer`/`io.Reader`. File methods do not accept contexts.

**v0.2.1:**

<!-- example: legacy -->
```go
func download(p bambu.Printer, path string, dst io.Writer) error {
	return p.DownloadFile(path, dst)
}
```

**v0.3.0:**

<!-- example: current -->
```go
func download(p bambu.Printer, path string, dst io.Writer) error {
	files := p.Files()
	if files == nil {
		return bambu.ErrFTPUnavailable
	}
	return files.Download(path, dst)
}
```

FTP setup remains optional. In the current implementation, failed FTP setup leaves `Files()` nil. Check it before calling any method; calling through nil does not return `ErrFTPUnavailable`. The example translates nil into that error for the caller.

## Refresh and raw G-code

`p.RequestUpdate(ctx)` remains callable through `bambu.Printer`. It sends a refresh request but does not wait for the resulting state.

`SendGcode(ctx, []string)` remains callable through `bambu.Printer`, including values returned by `Client.Add` and `Load`. Existing calls need no changes:

<!-- example: current -->
```go
func sendGcode(ctx context.Context, p bambu.Printer, lines []string) error {
	return p.SendGcode(ctx, lines)
}
```

Raw G-code does not use the fan/light setters' optimistic updates or command gates. Use the system setters when you need those behaviors.

## Remaining compatibility changes

- `State()`, `GcodeState`, and its constants are removed. Use `Print().Info()` and `PrintState` for print status. Temperatures and camera/network telemetry still have no public getters.
- Custom light-flashing timing remains unavailable.
- `ModelP1P` was inserted before `ModelP2S`. Persisted numeric values for `ModelP2S` and subsequent models need explicit migration.
- `Capability` changes from `uint8` to `uint64` with different flags. Old masks and `ModelInfo` are not compatible; `Printer` has no capability getter.
- Custom `Printer` implementations now need `Serial`, `Close`, `RequestUpdate`, `SendGcode`, `Print`, `Lights`, `AMS`, `Fans`, `HMS`, and `Files`.
- `CommandClient.Send` and decoder APIs reference internal types. External-module tests should use application-level adapters rather than import `internal` packages.

After migrating, compile your application and check command receipt separately from optimistic state in integration tests.
