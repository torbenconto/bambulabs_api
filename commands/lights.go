package commands

import (
	"fmt"

	"github.com/torbenconto/bambulabs_api/internal/mqtt"
)

type Light string

const (
	ChamberLight Light = "chamber_light"
	PartLight    Light = "part_light"
)

func (l Light) String() string {
	switch l {
	case ChamberLight:
		return "Chamber light"
	case PartLight:
		return "Part light"
	default:
		return "Unknown"
	}
}

type Mode string

const (
	Off      Mode = "off"
	On       Mode = "on"
	Flashing Mode = "flashing"
)

func (m Mode) String() string {
	switch m {
	case Off:
		return "Off"
	case On:
		return "On"
	case Flashing:
		return "Flashing"
	default:
		return "Unknown"
	}
}

type Lights struct {
	mqttClient *mqtt.Client
}

func CreateLightsInstance(mqttClient *mqtt.Client) *Lights {
	return &Lights{mqttClient: mqttClient}
}

func (l *Lights) setLight(light Light, mode Mode) error {
	// The fields led_on_time, led_off_time, loop_times, and interval_time are only used for mode "flashing" but are required nonetheless.
	command := mqtt.NewCommand(mqtt.System).
		AddCommandField("ledctrl").
		AddField("led_node", light).
		AddField("led_mode", mode)
	if mode == Flashing {
		command.AddField("led_on_time", 500).
			AddField("led_off_time", 500).
			AddField("loop_times", 1).
			AddField("interval_time", 1000)
	}

	if err := l.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	return nil
}

func (l *Lights) TurnOn(light Light) {
	l.setLight(light, On)
}

func (l *Lights) TurnOff(light Light) {
	l.setLight(light, Off)
}

func (l *Lights) Flash(light Light) {
	l.setLight(light, Flashing)
}

func (l *Lights) SetLightFlashing(light Light, mode Mode, onTime int, offTime int, loopTimes int, intervalTime int) error {
	command := mqtt.NewCommand(mqtt.System).
		AddCommandField("ledctrl").
		AddField("led_node", light).
		AddField("led_mode", mode).
		AddField("led_on_time", onTime).
		AddField("led_off_time", offTime).
		AddField("loop_times", loopTimes).
		AddField("interval_time", intervalTime)

	if err := l.mqttClient.Publish(command); err != nil {
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	return nil
}
