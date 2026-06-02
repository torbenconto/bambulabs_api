<div align="center">

<img src="./assets/Logo-boxed.svg" alt="Logo">

<h1>Bambulabs API Golang Library</h1>
</div>

> [!IMPORTANT]
> This library is still in development. Consider [starring the repository](https://docs.github.com/en/get-started/exploring-projects-on-github/saving-repositories-with-stars) to show your support.

This repository provides a **Golang** library to interface with **Bambulabs 3D printers** via network protocols. It allows easy integration of Bambulabs printers into your Go applications, providing access to printer data, control over printer features, and more.

This project does not support the bambulabs cloud api, but it's sister project [bambulabs_cloud_api](https://github.com/torbenconto/bambulabs_cloud_api) does.
<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=torbenconto/bambulabs_api,torbenconto/bambulabs_cloud_api&type=Date)](https://star-history.com/#torbenconto/bambulabs_api&torbenconto/bambulabs_cloud_api&Date)
</div>

## Table of Contents

- [Installation](#installation)
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
config := &bambulabs_api.Config{
    Host:         net.ParseIP("12.34.56.78"),
    Model:        bambulabs_api.ModelUnknown // Change to your printer's model
    AccessCode:   "accessCode"
    SerialNumber: "serialNumber"
}
new_client := bambulabs_api.NewClient(context.background())
printer, err := new_client.Add(config)
if err != nil {
    panic(err)
}
new_client.Close()
```

Adding the printer to the config will immediately initiate a connection attempt to the given IP via both MQTT and FTP. The client is an abstraction which holds all your current printers, and allows you easily manage their connections, data, commands, etc.

While the client stores your printer internally, you can also use the printer object returned by `Add()`, they are functionally identical. To retrieve a printer object from the client, you can use `Load()`.
Example:

```go
loaded_printer, err := new_client.Load("serialNumber")
if err != nil {
    panic(err)
}
```

You can iterate over stored printers using the `Range()` function:

```go
new_client.Range(func(p Printer) bool {
    fmt.Println(p.SerialNumber)
    return true
})
```

When you're done with the client, close it using the `Close()` method:

```go
new_client.Close()
```

### Whats the significance of the model variable?
The model variable within the `bambulabs_api.Config` struct is an extremely valuable variable, it's used to communicate your printer's constraints to the package. Ensure this variable is accurate or your program may throw an error or behave unexpectedly. If you're unsure of your model, or are using the program for basic compatibility testing, use `bambulabs_api.ModelUnknown`, this model ensures a conservative constraint list, maximizing compatibility.


## Development

### Current Status: FULL FUNCTIONALITY
The library is currently in a stable state with full functionality. All major features have been implemented and tested across various Bambulabs printer models. However, there may still be some edge cases or specific features that require further testing and validation. Contributions are welcome to improve functionality and expand coverage.

## Cool projects using this library
- [Bambu Lightshow](https://github.com/TrippHopkins/Bambu-Light-Show)

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
