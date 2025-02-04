package bambulabs_api

import (
	"fmt"

	"github.com/torbenconto/bambulabs_api/internal/camera"
	_light "github.com/torbenconto/bambulabs_api/light"
	"github.com/torbenconto/bambulabs_api/utils"

	"image/color"
	"strconv"
	"time"

	_commands "github.com/torbenconto/bambulabs_api/commands"
	"github.com/torbenconto/bambulabs_api/internal/ftp"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"

	"github.com/torbenconto/bambulabs_api/state"
)

type Printer struct {
	ipAddr     string
	accessCode string
	serial     string

	mqttClient *mqtt.Client
	ftpClient  *ftp.Client

	Camera *camera.Client
	Lights *_commands.Lights
	HMS    *_commands.HMS
	Prints *_commands.Prints
	Misc   *_commands.Misc
	FTP    *_commands.FTP
}

func NewPrinter(config *PrinterConfig) *Printer {
	mqttClient := mqtt.NewClient(&mqtt.ClientConfig{
		Host:       config.Host,
		Port:       8883,
		Serial:     config.SerialNumber,
		Username:   "bblp",
		AccessCode: config.AccessCode,
		Timeout:    5 * time.Second,
	})

	ftpClient := ftp.NewClient(&ftp.ClientConfig{
		Host:       config.Host,
		Port:       990,
		Username:   "bblp",
		AccessCode: config.AccessCode,
	})

	return &Printer{
		ipAddr:     config.Host,
		accessCode: config.AccessCode,
		serial:     config.SerialNumber,

		mqttClient: mqttClient,
		ftpClient:  ftpClient,
		Camera: camera.NewClient(&camera.ClientConfig{
			Hostname:   config.Host,
			AccessCode: config.AccessCode,
			Username:   "bblp",
			Port:       6000,
		}),
		Lights: _commands.CreateLightsInstance(mqttClient),
		HMS:    _commands.CreateHMSInstance(mqttClient),
		Prints: _commands.CreatePrintsInstance(mqttClient),
		Misc:   _commands.CreateMiscInstance(mqttClient),
		FTP:    _commands.CreateFTPInstance(ftpClient),
	}
}

// Connect connects to the underlying clients.
func (p *Printer) Connect() error {
	err := p.mqttClient.Connect()
	if err != nil {
		return fmt.Errorf("mqttClient.Connect() error %w", err)
	}

	err = p.ftpClient.Connect()
	if err != nil {
		return fmt.Errorf("ftpClient.Connect() error %w", err)
	}

	err = p.Camera.Connect()
	if err != nil {
		return fmt.Errorf("cameraClient.Connect() error %w", err)
	}

	return nil
}

// Disconnect disconnects from the underlying clients
func (p *Printer) Disconnect() error {
	p.mqttClient.Disconnect()

	err := p.ftpClient.Disconnect()
	if err != nil {
		return fmt.Errorf("ftpClient.Disconnect() error %w", err)
	}

	err = p.Camera.Disconnect()
	if err != nil {
		return fmt.Errorf("cameraClient.Disconnect() error %w", err)
	}

	return nil
}

func unsafeParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func unsafeParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// Data returns the current state of the printer as a Data struct.
// This function is currently working but problems exist with the underlying.
func (p *Printer) Data() (Data, error) {
	// Retrieve data from the MQTT client
	data := p.mqttClient.Data()

	// Initialize the final Data structure with basic fields
	final := Data{
		Ams:                     make([]Ams, 0),
		AmsExists:               data.Print.Ams.AmsExistBits == "1",
		BedTargetTemperature:    data.Print.BedTargetTemper,
		BedTemperature:          data.Print.BedTemper,
		AuxiliaryFanSpeed:       unsafeParseInt(data.Print.BigFan1Speed),
		ChamberFanSpeed:         unsafeParseInt(data.Print.BigFan2Speed),
		PartFanSpeed:            unsafeParseInt(data.Print.CoolingFanSpeed),
		HeatbreakFanSpeed:       unsafeParseInt(data.Print.HeatbreakFanSpeed),
		ChamberTemperature:      data.Print.ChamberTemper,
		GcodeFile:               data.Print.GcodeFile,
		GcodeFilePreparePercent: unsafeParseInt(data.Print.GcodeFilePreparePercent),
		GcodeState:              state.GcodeState(data.Print.GcodeState),
		PrintPercentDone:        data.Print.McPercent,
		PrintErrorCode:          data.Print.McPrintErrorCode,
		RemainingPrintTime:      data.Print.McRemainingTime,
		HMS:                     data.Print.HMS,
		NozzleDiameter:          data.Print.NozzleDiameter,
		NozzleTargetTemperature: data.Print.NozzleTargetTemper,
		NozzleTemperature:       data.Print.NozzleTemper,
		Sdcard:                  data.Print.Sdcard,
		WifiSignal:              data.Print.WifiSignal,
	}

	for _, light := range data.Print.LightsReport {
		final.LightReport = append(final.LightReport, LightReport{
			Node: _light.Light(light.Node),
			Mode: _light.Mode(light.Mode),
		})
	} // TODO: Transfer this to Commands

	colors := make([]color.RGBA, 0)
	for _, col := range data.Print.VtTray.Cols {
		if col == "" {
			colors = append(colors, color.RGBA{})
		} else {
			c, err := utils.ParseHexColorFast(col)
			if err != nil {
				return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
			}
			colors = append(colors, c)
		}
	}

	var trayColor = color.RGBA{}
	if data.Print.VtTray.TrayColor != "" {
		var err error
		trayColor, err = utils.ParseHexColorFast(data.Print.VtTray.TrayColor)
		if err != nil {
			return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
		}
	}

	final.VtTray = Tray{
		ID:                unsafeParseInt(data.Print.VtTray.ID),
		BedTemperature:    unsafeParseFloat(data.Print.VtTray.BedTemp),
		Colors:            colors,
		DryingTemperature: unsafeParseFloat(data.Print.VtTray.DryingTemp),
		DryingTime:        unsafeParseInt(data.Print.VtTray.DryingTime),
		NozzleTempMax:     unsafeParseFloat(data.Print.VtTray.NozzleTempMax),
		NozzleTempMin:     unsafeParseFloat(data.Print.VtTray.NozzleTempMin),
		TrayColor:         trayColor,
		TrayDiameter:      unsafeParseFloat(data.Print.VtTray.TrayDiameter),
		TraySubBrands:     data.Print.VtTray.TraySubBrands,
		TrayType:          data.Print.VtTray.TrayType,
		TrayWeight:        unsafeParseInt(data.Print.VtTray.TrayWeight),
	}

	// Process AMS data
	for _, ams := range data.Print.Ams.Ams {
		trays := make([]Tray, 0)

		// Process trays for each AMS
		for _, tray := range ams.Tray {
			colors := make([]color.RGBA, 0)

			// Process colors for each tray
			for _, col := range tray.Cols {
				if col == "" {
					colors = append(colors, color.RGBA{})
				}
				c, err := utils.ParseHexColorFast(col)
				if err != nil {
					return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
				}
				colors = append(colors, c)
			}

			var trayColor = color.RGBA{}
			if tray.TrayColor != "" {

				var err error
				trayColor, err = utils.ParseHexColorFast(tray.TrayColor)
				if err != nil {
					return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
				}
			}

			trays = append(trays, Tray{
				ID:                unsafeParseInt(tray.ID),
				BedTemperature:    unsafeParseFloat(tray.BedTemp),
				Colors:            colors,
				DryingTemperature: unsafeParseFloat(tray.DryingTemp),
				DryingTime:        unsafeParseInt(tray.DryingTime),
				NozzleTempMax:     unsafeParseFloat(tray.NozzleTempMax),
				NozzleTempMin:     unsafeParseFloat(tray.NozzleTempMin),
				TrayColor:         trayColor,
				TrayDiameter:      unsafeParseFloat(tray.TrayDiameter),
				TraySubBrands:     tray.TraySubBrands,
				TrayType:          tray.TrayType,
				TrayWeight:        unsafeParseInt(tray.TrayWeight),
			})
		}

		final.Ams = append(final.Ams, Ams{
			Humidity:    unsafeParseInt(ams.Humidity),
			ID:          unsafeParseInt(ams.ID),
			Temperature: unsafeParseFloat(ams.Temp),
			Trays:       trays,
		})
	}

	return final, nil
}

// region Get Data Functions

// GetSerial returns the serial number of the printer.
// This is used to identify the printer.
func (p *Printer) GetSerial() string {
	return p.serial
}

//TODO: Load/Unload filament, AMS stuff, set filament, set bed height

//endregion
