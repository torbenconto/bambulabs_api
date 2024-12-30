package bambulabs_api

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/torbenconto/bambulabs_api/ftp"
	"github.com/torbenconto/bambulabs_api/mqtt"
	"github.com/torbenconto/bambulabs_api/state"
	"github.com/torbenconto/bambulabs_api/types"
	"github.com/torbenconto/bambulabs_api/util"
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
func (p *Printer) Data() types.Data {
	return p.MQTTClient.Data()
}

// GetPrinterState gets the current state of the printer
func (p *Printer) GetPrinterState() state.GcodeState {
	return state.GetGcodeState(p.MQTTClient.Data().Print.GcodeState)
}

//region Publishing functions (Set)

func (p *Printer) LightOn() error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.System,
		Command: "light_mode",
		Param:   "on",
	}.JSON())
}

func (p *Printer) LightOff() error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.System,
		Command: "light_mode",
		Param:   "off",
	}.JSON())
}

func (p *Printer) StopPrint() error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "stop",
	}.JSON())
}

func (p *Printer) PausePrint() error {
	if p.GetPrinterState() == state.PAUSE {
		return nil
	}

	return p.MQTTClient.Publish(
		mqtt.Command{
			Type:    mqtt.Print,
			Command: "pause",
		}.JSON())
}

func (p *Printer) ResumePrint() error {
	if p.GetPrinterState() == state.RUNNING {
		return nil
	}

	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "resume",
	}.JSON())
}

// SendGcode sends gcode command lines in a list to the printer
func (p *Printer) SendGcode(gcode []string) error {
	for _, g := range gcode {
		if !util.IsValidGCode(g) {
			return fmt.Errorf("invalid gcode: %s", g)
		}

		err := p.MQTTClient.Publish(
			mqtt.Command{
				Type:    mqtt.Print,
				Command: "gcode_line",
				Param:   g,
			}.JSON())
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintGcodeFile prints a gcode file on the printer given an absolute path.
func (p *Printer) PrintGcodeFile(filePath string) error {
	return p.MQTTClient.Publish(
		mqtt.Command{
			Type:    mqtt.Print,
			Command: "gcode_file",
			Param:   filePath,
		}.JSON())
}

func (p *Printer) Print3mfFile(fileName string, plate int, useAms bool) error {
	// Probably doesent work. Need to check the correct format of the command
	//return p.MQTTClient.Publish(`{"print": {"command": "project_file", "param": "Metadata/plate_` + string(plate) + `.gcode", "subtask_name": ` + fileName + `, "use_ams": ` + strconv.FormatBool(useAms) + `"bed_leveling": true, "url": "ftp://"` + fileName + `, "bed_type": "auto", "flow_cali": true, "vibration_cali": true, "layer_inspect: true", "ams_mapping": [0]}}`)
	return errors.ErrUnsupported
}

// SetBedTemperature sets the bed temperature to a specified number in degrees Celcius using a gcode command.
func (p *Printer) SetBedTemperature(temperature int) error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "gcode_line",
		Param:   fmt.Sprintf("M140 S%d\n", temperature),
	}.JSON())
}

// SetBedTemperatureAndWaitUntilReached sets the bed temperature to a specified number in egrees Celcius and waits for it to be reached using a gcode command.
func (p *Printer) SetBedTemperatureAndWaitUntilReached(temperature int) error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "gcode_line",
		Param:   fmt.Sprintf("M190 S%d\n", temperature),
	}.JSON())
}

// SetFanSpeed sets the speed of fan to a speed between 0-255
func (p *Printer) SetFanSpeed(fan Fan, speed int) error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "gcode_line",
		Param:   fmt.Sprintf("M106 P%d S%d\n", fan, speed),
	}.JSON())
}

// SetNozzleTemperature sets the nozzle temperature to a specified number in degrees Celcius using a gcode command.
func (p *Printer) SetNozzleTemperature(temperature int) error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "gcode_line",
		Param:   fmt.Sprintf("M104 S%d\n", temperature),
	}.JSON())
}

// SetNozzleTemperatureAndWaitUntilReached sets the nozzle temperature to a specified number in degrees Celcius and waits for it to be reached using a gcode command.
func (p *Printer) SetNozzleTemperatureAndWaitUntilReached(temperature int) error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "gcode_line",
		Param:   fmt.Sprintf("M109 S%d\n", temperature),
	}.JSON())
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

	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "calibration",
		Param:   strconv.Itoa(bitmask),
	}.JSON())
}

// SetPrintSpeed sets the print speed to a specified speed of type Speed (Silent, Standard, Sport, Ludicrous)
func (p *Printer) SetPrintSpeed(speed Speed) error {
	return p.MQTTClient.Publish(mqtt.Command{
		Type:    mqtt.Print,
		Command: "print_speed",
		Param:   strconv.Itoa(int(speed)),
	}.JSON())
}

//TODO: Load/Unload filament, AMS stuff, set filament, set bed height

//endregion
