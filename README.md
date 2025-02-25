<div align="center">
    <picture>
      <source srcset="https://raw.githubusercontent.com/torbenconto/bambulabs_api/refs/heads/master/assets/Logo-light.svg" media="(prefers-color-scheme: dark)">
      <source srcset="https://raw.githubusercontent.com/torbenconto/bambulabs_api/refs/heads/master/assets/Logo-dark.svg" media="(prefers-color-scheme: light)">
      <img src="https://raw.githubusercontent.com/torbenconto/bambulabs_api/refs/heads/master/assets/Logo-dark.svg" alt="Logo">
    </picture>

<h1>Bambulabs API Golang Library</h1>
</div>

> [!IMPORTANT]
> This app is still in development. Consider [starring the repository](https://docs.github.com/en/get-started/exploring-projects-on-github/saving-repositories-with-stars) to show your support.

This repository provides a **Golang** library to interface with **Bambulabs 3D printers** via network protocols. It allows easy integration of Bambulabs printers into your Go applications, providing access to printer data, control over printer features, and more.

This project does not support the bambulabs cloud api, but it's sister project [bambulabs_cloud_api](https://github.com/torbenconto/bambulabs_cloud_api) does.
<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=torbenconto/bambulabs_api,torbenconto/bambulabs_cloud_api&type=Date)](https://star-history.com/#torbenconto/bambulabs_api&torbenconto/bambulabs_cloud_api&Date)
</div>

## Table of Contents

- [Installation](#installation)
- [Connecting to a Printer](#connecting-to-a-printer)
- [Basic Examples](#basic-examples)
- [Development](#development)
- [Contributing](#contributing)
- [Links & Resources](#links--resources)
- [License](#license)


## Installation

To install the Bambulabs API Golang library, use the `go get` command:

```bash
go get -u github.com/torbenconto/bambulabs_api
```


## Connecting to a Printer

To interact with a Bambulabs printer, you need the following details:

- **IP Address**: The local IP address of the printer.
- **Serial Number**: The unique serial number of the printer.
- **Access Code**: A local access code for authentication.

You can find the **IP Address** and **Access Code** in the printer’s network settings. Please refer to the guides below for more detailed instructions:

- [Find your printer's IP Address](https://intercom.help/octoeverywhere/en/articles/9034934-find-your-bambu-lab-printer-ip-address)
- [Find your printer's Access Code](https://intercom.help/octoeverywhere/en/articles/9028357-find-your-bambu-lab-printer-access-code)
- [Find your printer's Serial Number](https://wiki.bambulab.com/en/general/find-sn)

Once you have the necessary details, you can create and connect to the printer with the following code:

```go
// Replace the values below with the ones for your printer
config := &bambulabs_api.PrinterConfig{
    Host:           "192.168.1.200",
    AccessCode:   ACCESS_CODE,
    SerialNumber: SERIAL_NUMBER,
}
printer := bambulabs_api.NewPrinter(config)
err := printer.Connect()
if err != nil {
    panic(err)
}
```

## Managing Multiple Printers
The PrinterPool is a concurrent, thread-safe structure designed to manage multiple printers in a pool. It allows you to interact with the printers by serial number, retrieve their status or data, and perform operations on them, all while handling multiple printers concurrently.
To begin using the PrinterPool, first create an instance of it:
```go
pool := bambulabs_api.NewPrinterPool()
```
Next, add printers to the pool using the AddPrinter method and passing in a config:
```go
configs := []*PrinterConfig{
    {Host: "192.168.1.200", SerialNumber: "M123ALE29D", AccessCode: "ODJ2j3"},
    {Host: "192.168.1.201", SerialNumber: "M123ALE29E", AccessCode: "LDAdj3"},
}

for _, config := range configs {
    pool.AddPrinter(config)
}
```

Next, connect to the printers using the built-in ConnectAll method:
```go
err := pool.ConnectAll()
if err != nil {
    panic(err)
}
```

Once connected, you can interact with the printers in the pool using the various methods provided by the PrinterPool struct. For example, you can toggle the light on all printers in the pool:
```go
err := pool.ExecuteAll(func(printer *Printer) error {
	return printer.LightOn(light.ChamberLight)
})

if err != nil {
    panic(err)
}
```

You can also retrieve the data of all printers in the pool:
```go
data, err := pool.DataAll()
if err != nil {
    panic(err)
}

for _, printerData := range data {
    fmt.Println(printerData)
}
```

For operations on one printer in a pool, you can retrieve a printer by serial number using the At method:
```go
printer, err := pool.At("M123ALE29D")
if err != nil {
    panic(err)
}
```

Finally, you can disconnect from all printers in the pool using the DisconnectAll method:
```go
err := pool.DisconnectAll()
if err != nil {
    panic(err)
}
```

## Basic Examples

Here is a basic example of how to create a connection to a printer and interact with it. This example toggles the printer's light and retrieves some printer data:

```go
package main

import (
	"fmt"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/light"
	"time"
	"net"
)

func main() {
	config := &bambulabs_api.PrinterConfig{
		Host:           "192.168.1.200",
		AccessCode:   "00293KD0",
		SerialNumber: "AC1029391BH109",
	}

	// Create printer object
	printer := bambulabs_api.NewPrinter(config)

	// Connect to printer via MQTT
	err := printer.Connect()
	if err != nil {
		panic(err)
    }

	// Attempt to toggle light
	err = printer.LightOn(light.ChamberLight)
	if err != nil {
		panic(err)
	}
	
	for {
		time.Sleep(1 * time.Second)

        data, err := printer.Data()
		if err != nil {
			panic(err)
        }
		if !data.IsEmpty() {
			fmt.Println(data)
        }
    }
}
```

This example establishes a printer pool with two printers, connects to them, toggles the light on both printers, and retrieves their data:

```go
package main

import (
    "fmt"
    "github.com/torbenconto/bambulabs_api"
    "github.com/torbenconto/bambulabs_api/light"
    "net"
    "time"
)

func main() {
    pool := bambulabs_api.NewPrinterPool()

    configs := []*bambulabs_api.PrinterConfig{
        {Host: "192.168.1.200", SerialNumber: "M123ALE29D", AccessCode: "ODJ2j3"},
        {Host: "192.168.1.201", SerialNumber: "M123ALE29E", AccessCode: "LDAdj3"},
    }

    for _, config := range configs {
        pool.AddPrinter(config)
    }

    err := pool.ConnectAll()
    if err != nil {
        panic(err)
    }

    err = pool.ExecuteAll(func(printer *bambulabs_api.Printer) error {
		return printer.LightOn(light.ChamberLight)
	})

    if err != nil {
        panic(err)
    }

    for {
        time.Sleep(1 * time.Second)

        data, err := pool.DataAll()
        if err != nil {
            panic(err)
        }

        for _, printerData := range data {
            fmt.Println(printerData)
        }
    }
}
```

## Development

### Current Status: ALPHA

This library is in active development. While many features have been implemented, certain functions are not fully tested across all supported devices. Contributions are welcome to improve functionality and expand coverage.

## Contributing

We welcome contributions to improve this project! If you’d like to contribute, please follow these steps:

1. Fork the repository
2. Clone your fork
3. Create a new branch for your feature or bug fix
4. Write tests if applicable
5. Submit a pull request with a detailed description of your changes

Please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) file for more details on how to contribute.


## Links & Resources

- [Bambulab Official Website](https://www.bambulab.com)
- [Bambulab Wiki](https://wiki.bambulab.com)
- [Bambulab Support](https://support.bambulab.com)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

Feel free to join the community and connect with us for help, suggestions, or collaborations:  
[Join the Discord!](https://discord.gg/7wmQ6kGBef)
