package bambulabs_api

import (
	"fmt"
	_fan "github.com/torbenconto/bambulabs_api/fan"
	"github.com/torbenconto/bambulabs_api/internal/ftp"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	_light "github.com/torbenconto/bambulabs_api/light"
	_printspeed "github.com/torbenconto/bambulabs_api/printspeed"
	"image/color"
	"os"
	"strconv"
	"time"

	"github.com/torbenconto/bambulabs_api/state"
)

type Printer struct {
	ipAddr     string
	accessCode string
	serial     string

	mqttClient *mqtt.Client
	ftpClient  *ftp.Client
}

func NewPrinter(config *PrinterConfig) *Printer {
	return &Printer{
		ipAddr:     config.Host,
		accessCode: config.AccessCode,
		serial:     config.SerialNumber,

		mqttClient: mqtt.NewClient(&mqtt.ClientConfig{
			Host:       config.Host,
			Port:       8883,
			Serial:     config.SerialNumber,
			Username:   "bblp",
			AccessCode: config.AccessCode,
			Timeout:    5 * time.Second,
		}),
		ftpClient: ftp.NewClient(&ftp.ClientConfig{
			Host:       config.Host,
			Port:       990,
			Username:   "bblp",
			AccessCode: config.AccessCode,
		}),
	}
}

// Connect connects to the underlying clients.
func (p *Printer) Connect() error {
	err := p.ConnectMqtt()
	if err != nil {
		return fmt.Errorf("mqttClient.Connect() error %w", err)
	}

	err = p.ConnectFtp()
	if err != nil {
		return fmt.Errorf("ftpClient.Connect() error %w", err)
	}

	return nil
}

// ConnectMqtt only connects to the MQTT client.
func (p *Printer) ConnectMqtt() error {
	return p.mqttClient.Connect()
}

// ConnectFtp only connects to the FTP client.
func (p *Printer) ConnectFtp() error {
	return p.ftpClient.Connect()
}

// Disconnect disconnects from the underlying clients
func (p *Printer) Disconnect() error {
	p.mqttClient.Disconnect()
	if err := p.ftpClient.Disconnect(); err != nil {
		return fmt.Errorf("ftpClient.Disconnect() error %w", err)
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
// TODO: HMS
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
		NozzleDiameter:          data.Print.NozzleDiameter,
		NozzleTargetTemperature: data.Print.NozzleTargetTemper,
		NozzleTemperature:       data.Print.NozzleTemper,
		Sdcard:                  data.Print.Sdcard,
		WifiSignal:              data.Print.WifiSignal,
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
				c, err := parseHexColorFast(col)
				if err != nil {
					return Data{}, fmt.Errorf("parseHexColorFast() error %w", err)
				}
				colors = append(colors, c)
			}

			var trayColor = color.RGBA{}
			if tray.TrayColor != "" {

				var err error
				trayColor, err = parseHexColorFast(tray.TrayColor)
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

// GetPrinterState gets the current state of the printer.
// This function is currently working but problems exist with the underlying.
func (p *Printer) GetPrinterState() state.GcodeState {
	return state.GcodeState(p.mqttClient.Data().Print.GcodeState)
}

//region Publishing functions (Set)

// Light sets a given light to on (set=true) or off (set=false).
// TODO: Implement light flashing.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (p *Printer) Light(light _light.Light, set bool) error {
	// This light_mode is currently believed to be deprecated, leaving here commented in case it ends up being useful later.
	//command, err := mqtt.NewCommand(mqtt.System).AddCommandField("light_mode").AddParamField("on").JSON()
	//if err != nil {
	//	return err
	//}

	var mode string
	if set {
		mode = "on"
	} else {
		mode = "off"
	}

	// https://github.com/Doridian/OpenBambuAPI/blob/main/mqtt.md#systemledctrl
	command := mqtt.NewCommand(mqtt.System).AddCommandField("ledctrl").AddField("led_node", light).AddField("led_mode", mode)
	// Add fields only used for mode "flashing" but required nonetheless
	command.AddField("led_on_time", 500)
	command.AddField("led_off_time", 500)
	command.AddField("loop_times", 1)
	command.AddField("interval_time", 1000)

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	return nil
}

// StopPrint fully stops the current print job.
// Function works independently but problems exist with the underlying.
func (p *Printer) StopPrint() error {
	s := p.GetPrinterState()

	if s == state.IDLE || s == state.UNKNOWN {
		return fmt.Errorf("error pausing print: %s", s.String())
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("stop")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error stopping print: %w", err)
	}

	return nil
}

// PausePrint pauses the current print job.
// Function works independently but problems exist with the underlying.
func (p *Printer) PausePrint() error {
	s := p.GetPrinterState()

	if s == state.PAUSE || s == state.UNKNOWN {
		return fmt.Errorf("error pausing print: %s", s.String())
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("pause")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error pausing print: %w", err)
	}

	return nil
}

// ResumePrint resumes a paused print job.
// Function works independently but problems exist with the underlying.
func (p *Printer) ResumePrint() error {
	s := p.GetPrinterState()

	if s == state.RUNNING || s == state.UNKNOWN {
		return fmt.Errorf("error pausing print: %s", s.String())
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("resume")

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error resuming print: %w", err)
	}

	return nil
}

// SendGcode sends gcode command lines in a list to the printer.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (p *Printer) SendGcode(gcode []string) error {
	for _, g := range gcode {
		if !isValidGCode(g) {
			return fmt.Errorf("invalid gcode: %s", g)
		}

		command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(g)

		if err := p.mqttClient.Publish(command); err != nil {
			return fmt.Errorf("error sending gcode line %s: %w", g, err)
		}
	}
	return nil
}

// PrintGcodeFile prints a gcode file on the printer given an absolute path.
// This function is untested
func (p *Printer) PrintGcodeFile(filePath string) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_file").AddParamField(filePath)

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error printing gcode file %s: %w", filePath, err)
	}

	return nil
}

// Print3mfFile prints a ".gcode.3mf" file which resides on the printer. A file url (beginning with ftp:/// or file:///) should be passed in.
// You can upload a file through the ftp store function, then print it with this function using the url ftp:///[filename]. Make sure that it ends in .gcode or .gcode.3mf.
// The plate number should almost always be 1.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (p *Printer) Print3mfFile(fileUrl string, plate int, useAms bool, timelapse bool, calibrate bool, inspectLayers bool) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("project_file").AddParamField(fmt.Sprintf("Metadata/plate_%d.gcode", plate))

	command.AddField("project_id", "0")
	command.AddField("profile_id", "0")
	command.AddField("task_id", "0")
	command.AddField("subtask_id", "0")
	command.AddField("subtask_name", "")
	command.AddField("file", "")
	command.AddField("url", fileUrl)
	command.AddField("md5", "")
	command.AddField("timelapse", timelapse)
	command.AddField("bed_type", "auto")
	command.AddField("bed_levelling", calibrate)
	command.AddField("flow_cali", calibrate)
	command.AddField("vibration_cali", calibrate)
	command.AddField("layer_inspect", inspectLayers)
	command.AddField("ams_mapping", "")
	command.AddField("use_ams", useAms)

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error printing %s: %w", fileUrl, err)
	}

	return nil
}

// SetBedTemperature sets the bed temperature to a specified number in degrees Celcius using a gcode command.
func (p *Printer) SetBedTemperature(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M140 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting bed temperature: %w", err)
	}

	return nil
}

// SetBedTemperatureAndWaitUntilReached sets the bed temperature to a specified number in degrees Celsius and waits for it to be reached using a gcode command.
func (p *Printer) SetBedTemperatureAndWaitUntilReached(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M190 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting bed temperature and waiting for it to be reached: %w", err)
	}

	return nil
}

// SetFanSpeed sets the speed of fan to a speed between 0-255.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (p *Printer) SetFanSpeed(fan _fan.Fan, speed int) error {
	if speed < 0 || speed > 255 {
		return fmt.Errorf("invalid speed: %d; must be between 0 and 255", speed)
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M106 P%d S%d", fan, speed))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting fan speed: %w", err)
	}

	return nil
}

// SetNozzleTemperature sets the nozzle temperature to a specified number in degrees Celsius using a gcode command.
// This function is untested, but the underlying is working so it is likely to work.
func (p *Printer) SetNozzleTemperature(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M104 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting nozzle temperature: %w", err)
	}

	return nil
}

// SetNozzleTemperatureAndWaitUntilReached sets the nozzle temperature to a specified number in degrees Celsius and waits for it to be reached using a gcode command.
// This function is untested, but the underlying is working so it is likely to work.
func (p *Printer) SetNozzleTemperatureAndWaitUntilReached(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M109 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting nozzle temperature and waiting for it to be reached: %w", err)
	}

	return nil
}

// Calibrate runs the printer through a calibration process.
// This function is currently untested.
func (p *Printer) Calibrate(levelBed, vibrationCompensation, motorNoiseCancellation bool) error {
	bitmask := 0

	if levelBed {
		bitmask |= 1 << 1
	}
	if vibrationCompensation {
		bitmask |= 1 << 2
	}
	if motorNoiseCancellation {
		bitmask |= 1 << 3
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("calibration").AddParamField(strconv.Itoa(bitmask))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error calibrating: %w", err)
	}

	return nil
}

// SetPrintSpeed sets the print speed to a specified speed of type Speed (Silent, Standard, Sport, Ludicrous)
// This function is currently untested.
func (p *Printer) SetPrintSpeed(speed _printspeed.PrintSpeed) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("print_speed").AddParamField(speed)

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting print speed: %w", err)
	}

	return nil
}

//TODO: Load/Unload filament, AMS stuff, set filament, set bed height

//endregion

// region FTP functions

// StoreFile calls the underlying ftp client to store a file on the printer.
func (p *Printer) StoreFile(path string, file os.File) error {
	return p.ftpClient.StoreFile(path, file)
}

// ListDirectory calls the underlying ftp client to list the contents of a directory on the printer.
func (p *Printer) ListDirectory(path string) ([]os.FileInfo, error) {
	return p.ftpClient.ListDir(path)
}

// RetrieveFile calls the underlying ftp client to retrieve a file from the printer.
func (p *Printer) RetrieveFile(path string) (os.File, error) {
	return p.ftpClient.RetrieveFile(path)
}

// DeleteFile calls the underlying ftp client to delete a file from the printer.
func (p *Printer) DeleteFile(path string) error {
	return p.ftpClient.DeleteFile(path)
}

//endregion
