---
title: "Quickstart"
---

# Quickstart

This guide will walk you through the basics of using the Bambulabs API, including descriptions of some common use cases and examples. This guide is intended for beginners in the library with a basic understanding of programming and the go language. Those with prior library or 3d printing knowledge can also benefit from it.

The name `bambulabs_api` is the official name of this golang package, and has no affiliation with Bambulabs. It will be referred to as "the library" or "Bambulabs API" in this guide.

## Connection

The library connects to your printer over your local network (henceforth referred to as "LAN" or "local network"). It uses the MQTT protocol to access your printer's telemetry, allowing you to monitor and control your printer remotely, and the FTP protocol to browse, upload, and download files on the printer's storage (typically its SD card). Both connections are established automatically when a printer is added to the client — you don't need to manage them separately.

In order to connect to your printer, the library requires a couple pieces of information:

- Your printer's IP address (on the local network)
- Your printer's serial number
- Your printer's local access code

For more information on how to obtain these values, see the [README](../README.md).

Once you have these values, you can connect to your printer using the `bambulabs_api` library.

The library uses a central `Client` struct to manage connections and state. You can create a client using standard Go idioms. `NewClient` takes in a context for lifetime management.

```go
client, err := bambulabs_api.NewClient(context.Background())
```

Once you have obtained a `Client` instance, you can add your printer to the client by calling `Add` and passing in a `Config` struct with your printer's IP address, serial number, model, and local access code.

```go
printer, err := client.Add(bambulabs_api.Config{
    Host:         net.ParseIP("12.34.56.78"),
    SerialNumber: "ABC123",
    Model:        bambulabs_api.ModelUnknown,
    AccessCode:   "ACCESS_CODE",
})
```

`Add` establishes an MQTT connection to your printer, which is required for `Add` to succeed. It also attempts an FTP connection for file access; if the FTP connection fails (e.g. an unreachable port or misconfigured firewall), `Add` still succeeds, but any subsequent call to a file method (`ListFiles`, `DownloadFile`, `UploadFile`, `DeleteFile`) will return `bambulabs_api.ErrFTPUnavailable` until the printer is re-added.


The library requires the model of your printer to be specified in the `Model` field of the `Config` struct. This is required to determine which features are supported by your printer. Ensure this variable is accurate or your program may throw an error or behave unexpectedly. If you're unsure of your model, or are using the program for basic compatibility testing, use `bambulabs_api.ModelUnknown`, this model ensures a conservative constraint list and maximizes compatibility.

To see the list of supported models and their respective Model field values, see the [supported models documentation](supported_models.md).

`Config` also accepts optional `MQTTPort` and `FTPPort` fields if your printer uses non-default ports; if left unset they default to `8883` (MQTT over TLS) and `990` (FTP over implicit TLS) respectively.

Now that you have a `Printer` instance, you can interact with your printer using the various methods available. Not every method is available on every printer model, and not every method will be covered in this brief quickstart guide.

- Request an update and read the last-known state

```go
if err := printer.RequestUpdate(); err != nil {
    log.Printf("request update failed: %v", err)
}

if st, ok := printer.State(); ok {
    fmt.Printf("last state: %+v\n", st)
} else {
    fmt.Println("no state available yet")
}
```

- Control lights (models may not support every light)

```go
if err := printer.SetLight(bambulabs_api.ChamberLight, bambulabs_api.LightOn); err != nil {
    log.Printf("set light: %v", err)
}
```

- Set a fan speed

```go
// speed is 0-255
if err := printer.SetFan(bambulabs_api.ChamberFan, 255); err != nil {
    log.Printf("set fan: %v", err)
}
```

- Send raw G-code lines

```go
if err := printer.SendGcode([]string{"G28 ; home", "G1 X10 Y10 F600"}); err != nil {
    log.Printf("send gcode: %v", err)
}
```

## Files (FTP)

In addition to MQTT-based telemetry and control, the library exposes basic file operations over the printer's FTP connection. This is useful for listing, uploading, or downloading files. For example: 3MF/G-code files on the printer's SD card.

**Note:** FTP connectivity is optional. If it couldn't be established when the printer was added, file methods return `bambulabs_api.ErrFTPUnavailable` rather than failing printer setup entirely. Check for this error if you want to distinguish "not connected" from other failures:

```go
if err := printer.DeleteFile("/model.gcode"); errors.Is(err, bambulabs_api.ErrFTPUnavailable) {
    log.Println("file access unavailable for this printer")
}
```

- List files in a directory

```go
entries, err := printer.ListFiles("/")
if err != nil {
    log.Printf("list files: %v", err)
}
for _, e := range entries {
    fmt.Println(e.Name())
}
```

- Download a file

```go
f, err := os.Create("model.gcode")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

if err := printer.DownloadFile("/model.gcode", f); err != nil {
    log.Printf("download file: %v", err)
}
```

- Upload a file

```go
f, err := os.Open("model.gcode")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

if err := printer.UploadFile("/model.gcode", f); err != nil {
    log.Printf("upload file: %v", err)
}
```

- Delete a file

```go
if err := printer.DeleteFile("/model.gcode"); err != nil {
    log.Printf("delete file: %v", err)
}
```

## Managing multiple printers

- Iterate over all printers managed by the client

```go
client.Range(func(p bambulabs_api.Printer) bool {
    fmt.Println("printer:", p.Serial())
    return true // continue iteration
})
```

- Remove a printer or close the client

```go
// remove and close a single printer
if err := client.Remove("MY-PRINTER-123"); err != nil {
    log.Printf("remove: %v", err)
}

// close all printers and stop the client
if err := client.Close(); err != nil {
    log.Printf("client close: %v", err)
}
```
