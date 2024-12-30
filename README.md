<div align="center" style="margin-left: 45px;">
    <picture>
      <source srcset="assets/Logo-light.svg" media="(prefers-color-scheme: dark)">
      <source srcset="assets/Logo-dark.svg" media="(prefers-color-scheme: light)">
      <img src="assets/Logo-dark.svg" alt="Logo">
    </picture>

<h1>Bambulabs API Golang Library</h1>
</div>

> [!IMPORTANT]
> This app is still in development and no public release is available. Consider [starring the repository](https://docs.github.com/en/get-started/exploring-projects-on-github/saving-repositories-with-stars) to show your support.

This repository provides a **Golang** library to interface with **Bambulabs 3D printers** via network protocols. It allows easy integration of Bambulabs printers into your Go applications, providing access to printer data, control over printer features, and more.

---

## Table of Contents

- [Installation](#installation)
- [Connecting to a Printer](#connecting-to-a-printer)
- [Basic Examples](#basic-examples)
- [Development](#development)
- [Contributing](#contributing)
- [Links & Resources](#links--resources)

---

## Installation

To install the Bambulabs API Golang library, use the `go get` command:

```bash
go get -u github.com/torbenconto/bambulabs_api
```

---

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
printer := bambulabs_api.NewPrinter(IP, ACCESS_CODE, SERIAL_NUMBER)
err := printer.Connect()
if err != nil {
    panic(err)
}
```

---

## Basic Examples

Here is a basic example of how to create a connection to a printer and interact with it. This example toggles the printer's light and retrieves some printer data:

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

**Note**: The `Light` function in the above example is untested. Ensure to validate functionality according to your printer's firmware.

---

## Development

### Current Status: UNTESTED

This library is in active development. While many features have been implemented, certain functions are not fully tested across all supported devices. Contributions are welcome to improve functionality and expand coverage.

---

## Contributing

We welcome contributions to improve this project! If you’d like to contribute, please follow these steps:

1. Fork the repository
2. Clone your fork
3. Create a new branch for your feature or bug fix
4. Write tests if applicable
5. Submit a pull request with a detailed description of your changes

Please refer to the [CONTRIBUTING.md](CONTRIBUTING.md) file for more details on how to contribute.

---

## Links & Resources

- [Bambulab Official Website](https://www.bambulab.com)
- [Bambulab Wiki](https://wiki.bambulab.com)
- [Bambulab Support](https://support.bambulab.com)

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

Feel free to join the community and connect with us for help, suggestions, or collaborations:  
[Join the Discord!](https://discord.gg/7wmQ6kGBef)
