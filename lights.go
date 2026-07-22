package bambulabs_api

import (
	"context"
	"sync"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

// LightInfo represents the ID and current mode of a printer light.
type LightInfo struct {
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
	mu     sync.RWMutex
	lights map[LightID]LightInfo // current state

	commandClient CommandClient
}

func NewLightSystem(commandClient CommandClient) *LightSystem {
	return &LightSystem{
		lights:        make(map[LightID]LightInfo, 6),
		commandClient: commandClient,
	}
}

// Get returns the last known state of the given light, as reported by the
// printer. It returns [ErrLightNotAvalible] if the printer hasn't reported
// this light yet.
func (l *LightSystem) Get(id LightID) (LightInfo, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	lt, ok := l.lights[id]
	if !ok {
		return LightInfo{}, ErrLightNotAvalible
	}

	return lt, nil
}

func (l *LightSystem) Set(ctx context.Context, id LightID, mode LightMode) error {
	if _, err := l.Get(id); err != nil {
		return err
	}

	return l.commandClient.Send(ctx, newLightCommand(id, mode, DefaultLightFlashingConfig()))
}

// apply records a light state reported by the printer. Called by
// [LightDecoder] while holding the printer's decode lock.
func (l *LightSystem) apply(id LightID, mode LightMode) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lights[id] = LightInfo{ID: id, Mode: mode}
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

type LightDecoder struct{}

func NewLightDecoder() *LightDecoder {
	return &LightDecoder{}
}

func (l *LightDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil {
		return
	}

	for _, rawLight := range report.Print.LightsReport {
		p.Lights.apply(LightID(rawLight.Node), LightMode(rawLight.Mode))
	}
}
