package bambulabs_api

import (
	"errors"
	"fmt"
	_fan "github.com/torbenconto/bambulabs_api/fan"
	_light "github.com/torbenconto/bambulabs_api/light"
	_speed "github.com/torbenconto/bambulabs_api/speed"
	"net"
	"strconv"
	"time"

	"github.com/torbenconto/bambulabs_api/ftp"
	"github.com/torbenconto/bambulabs_api/mqtt"
	"github.com/torbenconto/bambulabs_api/state"
)

type Printer struct {
	ipAddr     net.IP
	accessCode string
	serial     string

	MQTTClient *mqtt.Client
	FTPClient  *ftp.Client
}

func NewPrinter(ipAddr net.IP, accessCode, serial string) *Printer {
	return &Printer{
		ipAddr:     ipAddr,
		accessCode: accessCode,
		serial:     serial,

		MQTTClient: mqtt.NewClient(&mqtt.ClientConfig{
			Host:       ipAddr,
			Port:       8883,
			Serial:     serial,
			Username:   "bblp",
			AccessCode: accessCode,

			Timeout: 5 * time.Second,
		}),
		FTPClient: ftp.NewClient(&ftp.ClientConfig{
			Host:       ipAddr,
			Port:       990,
			Username:   "bblp",
			AccessCode: accessCode,
		}),
	}
}

func (p *Printer) Connect() error {
	err := p.MQTTClient.Connect()
	if err != nil {
		return fmt.Errorf("MQTTClient.Connect() error %w", err)
	}

	err = p.FTPClient.Connect()
	if err != nil {
		return fmt.Errorf("FTPClient.Connect() error %w", err)
	}

	return nil
}

func (p *Printer) Disconnect() error {
	p.MQTTClient.Disconnect()
	return p.FTPClient.Disconnect()
}

// Data returns the current state of the printer as a Data struct
func (p *Printer) Data() Data {
	return p.MQTTClient.Data()
}

// GetPrinterState gets the current state of the printer
func (p *Printer) GetPrinterState() state.GcodeState {
	return state.GetGcodeState(p.MQTTClient.Data().Print.GcodeState)
}

//region Publishing functions (Set)

// Light sets a given light to on (set=true) or off (set=false)
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

	return p.MQTTClient.Publish(command)
}

func (p *Printer) StopPrint() error {
	if p.GetPrinterState() == state.IDLE {
		return nil
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("stop")

	return p.MQTTClient.Publish(command)
}

func (p *Printer) PausePrint() error {
	if p.GetPrinterState() == state.PAUSE {
		return nil
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("pause")

	return p.MQTTClient.Publish(command)
}

func (p *Printer) ResumePrint() error {
	if p.GetPrinterState() == state.RUNNING {
		return nil
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("resume")

	return p.MQTTClient.Publish(command)
}

// SendGcode sends gcode command lines in a list to the printer
func (p *Printer) SendGcode(gcode []string) error {
	for _, g := range gcode {
		if !isValidGCode(g) {
			return fmt.Errorf("invalid gcode: %s", g)
		}

		command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(g)

		err := p.MQTTClient.Publish(command)

		if err != nil {
			return err
		}
	}
	return nil
}

// PrintGcodeFile prints a gcode file on the printer given an absolute path.
func (p *Printer) PrintGcodeFile(filePath string) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_file").AddParamField(filePath)

	return p.MQTTClient.Publish(command)
}

func (p *Printer) Print3mfFile(fileName string, plate int, useAms bool) error {
	// Probably doesent work. Need to check the correct format of the command
	//return p.MQTTClient.Publish(`{"print": {"command": "project_file", "param": "Metadata/plate_` + string(plate) + `.gcode", "subtask_name": ` + fileName + `, "use_ams": ` + strconv.FormatBool(useAms) + `"bed_leveling": true, "url": "ftp://"` + fileName + `, "bed_type": "auto", "flow_cali": true, "vibration_cali": true, "layer_inspect: true", "ams_mapping": [0]}}`)
	return errors.ErrUnsupported
}

// SetBedTemperature sets the bed temperature to a specified number in degrees Celcius using a gcode command.
func (p *Printer) SetBedTemperature(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M140 S%d", temperature))

	return p.MQTTClient.Publish(command)
}

// SetBedTemperatureAndWaitUntilReached sets the bed temperature to a specified number in degrees Celsius and waits for it to be reached using a gcode command.
func (p *Printer) SetBedTemperatureAndWaitUntilReached(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M190 S%d", temperature))

	return p.MQTTClient.Publish(command)
}

// SetFanSpeed sets the speed of fan to a speed between 0-255
func (p *Printer) SetFanSpeed(fan _fan.Fan, speed int) error {
	if speed < 0 || speed > 255 {
		return errors.New("speed must be between 0 and 255")
	}

	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M106 P%d S%d", fan, speed))

	return p.MQTTClient.Publish(command)
}

// SetNozzleTemperature sets the nozzle temperature to a specified number in degrees Celsius using a gcode command.
func (p *Printer) SetNozzleTemperature(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M104 S%d", temperature))

	return p.MQTTClient.Publish(command)
}

// SetNozzleTemperatureAndWaitUntilReached sets the nozzle temperature to a specified number in degrees Celsius and waits for it to be reached using a gcode command.
func (p *Printer) SetNozzleTemperatureAndWaitUntilReached(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M109 S%d", temperature))

	return p.MQTTClient.Publish(command)
}

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

	return p.MQTTClient.Publish(command)
}

// SetPrintSpeed sets the print speed to a specified speed of type Speed (Silent, Standard, Sport, Ludicrous)
func (p *Printer) SetPrintSpeed(speed _printspeed.PrintSpeed) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("print_speed").AddParamField(speed)

	return p.MQTTClient.Publish(command)
}

//TODO: Load/Unload filament, AMS stuff, set filament, set bed height

//endregion
