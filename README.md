# Bambulabs API

Golang library for interfacing with bambulabs printers.

[Join the Discord!](https://discord.gg/7wmQ6kGBef)

## Installation
```
    go get -u github.com/torbenconto/bambulabs_api
```

## Connecting to a printer
Making a connection to a printer requires 3 parameters:
- The local IP address of the printer
- The serial number of the printer
- The printer's local access code

The IP address and local access code can be found in the respective printers network settings.

[IP Guide](https://intercom.help/octoeverywhere/en/articles/9034934-find-your-bambu-lab-printer-ip-address)

[Access Code Guide](https://intercom.help/octoeverywhere/en/articles/9028357-find-your-bambu-lab-printer-access-code)


Finding the serial number is a little more complex and thus a few methods to do so on different printers can be found [here](https://wiki.bambulab.com/en/general/find-sn)

Once you have these parameters, you can create a printer object through 
```go
printer := bambulabs_api.NewPrinter(IP, ACCESS_CODE, SERIAL_NUMBER)
```
and connect to it via
```
printer.Connect()
```

## Basic Examples

```go
package main

import (
	"fmt"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/light"
	"net"
)

func main() {
	// Printer local IP 
	printerIp := net.IPv4(192, 168, 1, 200)
	// Printer serial number
	printerSerialNumber := "AC1029391BH109"
	// Printer access code
	printerAccessCode := "00293KD0"

	// Create printer object
	printer := bambulabs_api.NewPrinter(printerIp, printerAccessCode, printerSerialNumber)

	// Connect to printer via MQTT
	err := printer.Connect()
	if err != nil {
		panic(err)
    }

	// Attempt to toggle light (UNTESTED FUNCTION)
	err = printer.Light(light.ChamberLight, true)
	if err != nil {
		panic(err)
	}

	// Retrieve printer data
	data := printer.Data()
	fmt.Println(data)
}

```

## Development
current status: UNTESTED

