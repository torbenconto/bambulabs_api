package printer

import (
	"bambulabs-api/gcode_state"
	"bambulabs-api/mqtt"
	"bambulabs-api/print_state"
	"bambulabs-api/util"
	"fmt"
	"net"
	"time"
)

type Printer struct {
	ipAddr     net.IP
	accessCode string
	serial     string

	MQTTClient *mqtt.Client
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

			Timeout: 250 * time.Millisecond,
		}),
	}
}

func (p *Printer) Connect() error {
	return p.MQTTClient.Connect()
}

func (p *Printer) Disconnect() {
	p.MQTTClient.Disconnect()
}

func (p *Printer) update() error {
	return p.MQTTClient.Publish(mqtt.PushAll)
}

//region Status functions (Get)

/* TODO: review this functions return type */
// GetPrinterState gets the current state of the printer
func (p *Printer) GetPrinterState() gcode_state.GcodeState {
	return gcode_state.GetGcodeState(p.MQTTClient.Data()["gcode_state"].(int))
}

// GetPrintStatus gets the current state of the print
func (p *Printer) GetPrintStatus() print_state.PrintState {
	return print_state.GetPrinterState(p.MQTTClient.Data()["stg_cur"].(int))
}

// GetFileName gets the file name of the current or last print
func (p *Printer) GetFileName() string {
	return p.MQTTClient.Data()["gcode_file"].(string)
}

// GetPrintSpeed gets the current set print speed as a percentage
func (p *Printer) GetPrintSpeed() int {
	return p.MQTTClient.Data()["gcode_speed"].(int)
}

/* TODO: review this functions return type */
// GetLightState gets the state of the cabin light and returns either "on" "off" or "unknown"
func (p *Printer) GetLightState() string {
	d := p.MQTTClient.Data()["lights_report"].([]map[string]string)
	if len(d) == 0 {
		return "unknown"
	}

	return d[0]["mode"]
}

// GetBedTemperature retrieves the current bed temperature of the printer
func (p *Printer) GetBedTemperature() float64 {
	return p.MQTTClient.Data()["bed_temper"].(float64)
}

// GetBedTemperatureTarget retrieves the target bed temperature of the printer
func (p *Printer) GetBedTemperatureTarget() float64 {
	return p.MQTTClient.Data()["bed_target_temper"].(float64)
}

// GetNozzleTemperature retrieves the current nozzle temperature of the printer
func (p *Printer) GetNozzleTemperature() float64 {
	return p.MQTTClient.Data()["nozzle_temper"].(float64)
}

// GetNozzleTemperatureTarget retrieves the target nozzle temperature of the printer
func (p *Printer) GetNozzleTemperatureTarget() float64 {
	return p.MQTTClient.Data()["nozzle_target_temper"].(float64)
}

// CurrentLayerNum retrieves the current layer number of the print job
func (p *Printer) CurrentLayerNum() int {
	return p.MQTTClient.Data()["layer_num"].(int)
}

// TotalLayerNum retrieves the total layer count of the print job
func (p *Printer) TotalLayerNum() int {
	return p.MQTTClient.Data()["total_layer_num"].(int)
}

// GcodeFilePreparePercentage retrieves the percentage of the gcode file preparation
func (p *Printer) GcodeFilePreparePercentage() int {
	return p.MQTTClient.Data()["gcode_file_prepare_percent"].(int)
}

// NozzleDiameter retrieves the nozzle diameter currently registered to the printer
func (p *Printer) NozzleDiameter() float64 {
	return float64(p.MQTTClient.Data()["nozzle_diameter"].(float64))
}

// NozzleType retrieves the type of nozzle currently registered to the printer
func (p *Printer) NozzleType() Nozzle {
	// Assuming "NozzleType" is a type you define
	return Nozzle(p.MQTTClient.Data()["nozzle_type"].(string))
}

//endregion

//region Publishing functions (Set)

func (p *Printer) LightOn() error {
	return p.MQTTClient.Publish(mqtt.LightOn)
}

func (p *Printer) LightOff() error {
	return p.MQTTClient.Publish(mqtt.LightOff)
}

func (p *Printer) StopPrint() error {
	return p.MQTTClient.Publish(mqtt.Stop)
}

func (p *Printer) PausePrint() error {
	if p.GetPrinterState() == gcode_state.PAUSE {
		return nil
	}

	return p.MQTTClient.Publish(mqtt.Pause)
}

func (p *Printer) ResumePrint() error {
	if p.GetPrinterState() == gcode_state.RUNNING {
		return nil
	}

	return p.MQTTClient.Publish(mqtt.Resume)
}

// SendGcode sends gcode command lines in a list to the printer
func (p *Printer) SendGcode(gcode []string) error {
	for _, g := range gcode {
		if !util.IsValidGCode(g) {
			return fmt.Errorf("invalid gcode: %s", g)
		}

		err := p.MQTTClient.Publish(fmt.Sprintf(mqtt.GcodeTemplate, g))
		if err != nil {
			return err
		}
	}
	return nil
}

// SetBedTemperature sets the bed temperature to a specified number in degrees Celcius using a gcode command.
func (p *Printer) SetBedTemperature(temperature int) error {
	return p.MQTTClient.Publish(fmt.Sprintf(mqtt.GcodeTemplate, fmt.Sprintf("M140 S%d\n", temperature)))
}

// SetBedTemperatureAndWaitUntilReached sets the bed temperature to a specified number in egrees Celcius and waits for it to be reached using a gcode command.
func (p *Printer) SetBedTemperatureAndWaitUntilReached(temperature int) error {
	return p.MQTTClient.Publish(fmt.Sprintf(mqtt.GcodeTemplate, fmt.Sprintf("M190 S%d\n", temperature)))
}

// SetFanSpeed sets the speed of fan to a speed between 0-255
func (p *Printer) SetFanSpeed(fan Fan, speed int) error {
	return p.MQTTClient.Publish(fmt.Sprintf(mqtt.GcodeTemplate, fmt.Sprintf("M106 P%d S%d\n", fan, speed)))
}

// SetNozzleTemperature sets the nozzle temperature to a specified number in degrees Celcius using a gcode command.
func (p *Printer) SetNozzleTemperature(temperature int) error {
	return p.MQTTClient.Publish(fmt.Sprintf(mqtt.GcodeTemplate, fmt.Sprintf("M104 S%d\n", temperature)))
}

// SetNozzleTemperatureAndWaitUntilReached sets the nozzle temperature to a specified number in degrees Celcius and waits for it to be reached using a gcode command.
func (p *Printer) SetNozzleTemperatureAndWaitUntilReached(temperature int) error {
	return p.MQTTClient.Publish(fmt.Sprintf(mqtt.GcodeTemplate, fmt.Sprintf("M104 S%d\n", temperature)))
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

	return p.MQTTClient.Publish(fmt.Sprintf(mqtt.CalibrationTemplate, bitmask))
}

//TODO: Load/Unload filament, AMS stuff, set filament, set bed height

//endregion
