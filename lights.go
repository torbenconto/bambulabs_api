package bambulabs_api

import (
	"context"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type light struct {
	ID   LightID
	Mode LightMode
}

type LightID string

const (
	ChamberLight  LightID = "chamber_light"
	ChamberLight2 LightID = "chamber_light2"
	WorkLight     LightID = "work_light"
	HeatbedLight  LightID = "heatbed_light"
)

type LightMode string

const (
	LightModeOn       LightMode = "on"
	LightModeOff      LightMode = "off"
	LightModeFlashing LightMode = "flashing"
)

// LightFlashingConfig controls the timing of [LightModeFlashing].
type LightFlashingConfig struct {
	OnTime       time.Duration
	OffTime      time.Duration
	LoopTimes    int
	IntervalTime time.Duration
}

// DefaultLightFlashingConfig returns the flashing configuration.
func DefaultLightFlashingConfig() LightFlashingConfig {
	return LightFlashingConfig{
		OnTime:       500 * time.Millisecond,
		OffTime:      500 * time.Millisecond,
		LoopTimes:    1,
		IntervalTime: time.Second,
	}
}

type LightSystem struct {
	lights        map[LightID]light // current state
	commandClient CommandClient
}

func NewLightSystem(commandClient CommandClient) *LightSystem {
	return &LightSystem{
		lights:        make(map[LightID]light, 6),
		commandClient: commandClient,
	}
}

func (l *LightSystem) Set(ctx context.Context, light LightID, mode LightMode) error {
	if _, err := l.get(light); err != nil {
		return err
	}

	return l.commandClient.Send(ctx, newLightCommand(light, mode, DefaultLightFlashingConfig()))
}

func (l *LightSystem) get(id LightID) (light, error) {
	if light, ok := l.lights[id]; ok {
		return light, nil
	}

	return light{}, ErrLightNotAvalible
}

func newLightCommand(light LightID, mode LightMode, cfg LightFlashingConfig) *protocol.Command {
	return protocol.NewCommand(protocol.System).
		WithCommand("ledctrl").
		Set("led_node", light).
		Set("led_mode", mode).
		Set("led_on_time", cfg.OnTime.Milliseconds()).
		Set("led_off_time", cfg.OffTime.Milliseconds()).
		Set("loop_times", cfg.LoopTimes).
		Set("interval_time", cfg.IntervalTime.Milliseconds())
}

type LightDecoder struct {
}

func NewLightDecoder() *LightDecoder {
	return &LightDecoder{}
}

func (l *LightDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil {
		return
	}

	lightReport := report.Print.LightsReport
	for _, rawLight := range lightReport {
		mode := LightMode(rawLight.Mode)
		id := LightID(rawLight.Node)

		p.Lights.lights[id] = light{
			ID:   id,
			Mode: mode,
		}
	}
}
