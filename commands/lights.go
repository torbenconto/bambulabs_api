package commands

import (
	"fmt"

	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	_light "github.com/torbenconto/bambulabs_api/light"
)

type Lights struct {
	mqttClient *mqtt.Client
}

func CreateLightsInstance(mqttClient *mqtt.Client) *Lights {
	return &Lights{mqttClient: mqttClient}
}

func (l *Lights) setLight(light _light.Light, mode _light.Mode) error {
	// The fields led_on_time, led_off_time, loop_times, and interval_time are only used for mode "flashing" but are required nonetheless.
	command := mqtt.NewCommand(mqtt.System).
		AddCommandField("ledctrl").
		AddField("led_node", light).
		AddField("led_mode", mode)
	if mode == _light.Flashing {
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

func (l *Lights) TurnOn(light _light.Light) {
	l.setLight(light, _light.On)
}

func (l *Lights) TurnOff(light _light.Light) {
	l.setLight(light, _light.Off)
}

func (l *Lights) Flash(light _light.Light) {
	l.setLight(light, _light.Flashing)
}

func (l *Lights) SetLightFlashing(light _light.Light, mode _light.Mode, onTime int, offTime int, loopTimes int, intervalTime int) error {
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
