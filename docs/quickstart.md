---
title: "Quickstart"
---

# Quickstart

Use Go 1.26 and import `github.com/torbenconto/bambulabs_api`. You need the printer's LAN IP address, serial number, and access code. See the [README](../README.md) for where to find them.

## Connect and read print status

This complete example connects, reads the latest print information, and closes the client. Replace the configuration values with your printer's details.

<!-- example: program -->
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	bambu "github.com/torbenconto/bambulabs_api"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	client := bambu.NewClient(ctx)
	defer client.Close()

	p, err := client.Add(&bambu.Config{
		Host:         net.ParseIP("192.168.1.50"),
		SerialNumber: "PRINTER_SERIAL",
		AccessCode:   "ACCESS_CODE",
		Model:        bambu.ModelA1,
	})
	if err != nil {
		return err
	}

	info := p.Print().Info()
	fmt.Printf("%s: %s, %d%%, layer %d/%d, remaining %s\n",
		info.State, info.Name, info.ProgressPercent,
		info.CurrentLayer, info.TotalLayers, info.Remaining)
	return nil
}
```

`Add` takes a config pointer and waits for the initial print report. MQTT is required; FTP is optional. Ports default to 8883 and 990. Select the actual model for decoding defaults, or `ModelUnknown` if it is unknown.

The client context controls the entire connection lifetime. Keep the client open while using its printers. `client.Close()` closes all printers it owns.

## Read state

Getters read local state without network requests:

| System | Read |
| --- | --- |
| Print | `p.Print().Info()` returns state, job metadata, progress, layers, and remaining time. |
| Fans | `p.Fans().Get(bambu.PartCoolingFan)` returns `FanInfo` and an error. |
| Lights | `p.Lights().Get(bambu.ChamberLight)` returns `LightInfo` and an error. |
| AMS | `p.AMS().Units()`, `p.AMS().Get(unitID)`, `p.AMS().ExternalTray()`. |
| HMS | `p.HMS().Errors()` returns an independent error slice. |

`Print().Info()` returns one value, with no availability flag. Print reports replace the entire snapshot. See [print status](print.md) for field definitions.

The following examples are functions using the same `bambu` import alias. Add their standard-library imports as needed.

<!-- example: current -->
```go
func showHealth(p bambu.Printer) {
	for _, issue := range p.HMS().Errors() {
		fmt.Println(issue.GetCode(), issue.Error())
	}
}

func showFilament(p bambu.Printer) {
	for _, unit := range p.AMS().Units() {
		for _, tray := range unit.Trays {
			fmt.Printf("AMS %d, slot %d: %s (%d%%)\n",
				unit.ID, tray.Slot,
				tray.Filament.Material, tray.Filament.RemainingPercent)
		}
	}
	fmt.Println("External spool:", p.AMS().ExternalTray().Filament.Material)
}
```

AMS IDs are reported identifiers, not slice indexes. Treat AMS getter results as read-only.

## Control fans and lights

Commands accept a context. If it has no deadline, MQTT operations use a ten-second timeout.

<!-- example: current -->
```go
func setCooling(ctx context.Context, p bambu.Printer) error {
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return p.Fans().Set(opCtx, bambu.PartCoolingFan, 50)
}

func turnOnLight(ctx context.Context, p bambu.Printer) error {
	return p.Lights().Set(ctx, bambu.ChamberLight, bambu.LightModeOn)
}
```

Fan speeds are percentages from 0 to 100, rounded to the nearest ten. Light modes are `LightModeOn`, `LightModeOff`, and `LightModeFlashing`.

Fan/light getters and setters return `ErrFanUnavailable` or `ErrLightUnavailable` when the device has not been reported. Use `errors.Is` to check these errors.

Setters update local fan/light state immediately. A failed send rolls back the value unless a report has superseded it. A matching getter value does not confirm printer execution.

## Refresh and G-code

`RequestUpdate` sends a refresh request; it does not wait for the response. `SendGcode` sends the supplied lines without updating local state optimistically.

<!-- example: current -->
```go
func refresh(ctx context.Context, p bambu.Printer) error {
	return p.RequestUpdate(ctx)
}

func sendLines(ctx context.Context, p bambu.Printer, lines []string) error {
	return p.SendGcode(ctx, lines)
}
```

## Files

Check `p.Files()` before using it. It is nil when FTP setup failed. File methods do not accept contexts.

| Operation | Call |
| --- | --- |
| List | `files.List(path)` returns `[]os.FileInfo`. |
| Download | `files.Download(path, writer)` writes to an `io.Writer`. |
| Upload | `files.Upload(path, reader)` reads from an `io.Reader`. |
| Delete | `files.Delete(path)` removes the remote file. |

<!-- example: current -->
```go
func listFiles(p bambu.Printer) error {
	files := p.Files()
	if files == nil {
		return bambu.ErrFTPUnavailable
	}
	entries, err := files.List("/")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fmt.Println(entry.Name(), entry.Size())
	}
	return nil
}

func downloadFile(p bambu.Printer, remotePath, localPath string) error {
	files := p.Files()
	if files == nil {
		return bambu.ErrFTPUnavailable
	}
	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	return files.Download(remotePath, dst)
}
```

## Manage printers

`client.Load(serial)` returns an existing printer. `client.Remove(serial)` removes it and closes its connection. `client.Range` visits the managed printers:

<!-- example: current -->
```go
func showPrinters(client *bambu.Client) {
	client.Range(func(p bambu.Printer) bool {
		fmt.Println(p.Serial(), p.Print().Info().State)
		return true
	})
}
```

For API changes from v0.2.1, see the [migration guide](migration-v0.3.0.md).
