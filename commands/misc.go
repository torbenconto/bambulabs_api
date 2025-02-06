package commands

import (
	"fmt"
	"strconv"

	"github.com/torbenconto/bambulabs_api/fan"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/utils"
)

type Misc struct {
	mqttClient *mqtt.Client
}

func CreateMiscInstance(mqttClient *mqtt.Client) *Misc {
	return &Misc{mqttClient: mqttClient}
}

// SendGcode sends gcode command lines in a list to the printer.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (p *Misc) SendGcode(gcode []string) error {
	for _, g := range gcode {
		if !utils.IsValidGCode(g) {
			return fmt.Errorf("invalid gcode: %s", g)
		}

		command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(g)

		if err := p.mqttClient.Publish(command); err != nil {
			return fmt.Errorf("error sending gcode line %s: %w", g, err)
		}
	}
	return nil
}

// SetBedTemperature sets the bed temperature to a specified number in degrees Celcius using a gcode command.
func (p *Misc) SetBedTemperature(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M140 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting bed temperature: %w", err)
	}

	return nil
}

// SetBedTemperatureAndWaitUntilReached sets the bed temperature to a specified number in degrees Celsius and waits for it to be reached using a gcode command.
func (p *Misc) SetBedTemperatureAndWaitUntilReached(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M190 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting bed temperature and waiting for it to be reached: %w", err)
	}

	return nil
}

// SetFanSpeed sets the speed of fan to a speed between 0-255.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (p *Misc) SetFanSpeed(fan fan.Fan, speed int) error {
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
func (p *Misc) SetNozzleTemperature(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M104 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting nozzle temperature: %w", err)
	}

	return nil
}

// SetNozzleTemperatureAndWaitUntilReached sets the nozzle temperature to a specified number in degrees Celsius and waits for it to be reached using a gcode command.
// This function is untested, but the underlying is working so it is likely to work.
func (p *Misc) SetNozzleTemperatureAndWaitUntilReached(temperature int) error {
	command := mqtt.NewCommand(mqtt.Print).AddCommandField("gcode_line").AddParamField(fmt.Sprintf("M109 S%d", temperature))

	if err := p.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting nozzle temperature and waiting for it to be reached: %w", err)
	}

	return nil
}

// Calibrate runs the printer through a calibration process.
// This function is currently untested.
func (p *Misc) Calibrate(levelBed, vibrationCompensation, motorNoiseCancellation bool) error {
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
