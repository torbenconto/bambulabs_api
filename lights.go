package bambulabs_api

import (
	"context"
	"sync"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

// LightInfo represents the ID and current mode of a printer light.
type LightInfo struct {
	Light Light
	Mode  LightMode
}

type Light string

const (
	ChamberLight  Light = "chamber_light"
	ChamberLight2 Light = "chamber_light2"
	WorkLight     Light = "work_light"
	// HeatbedLight  Light = "heatbed_light" unclear
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
	sendGate  chan struct{}
	revisions map[Light]uint64
	mu        sync.RWMutex
	lights    map[Light]LightInfo

	commandClient CommandClient
}

func NewLightSystem(commandClient CommandClient) *LightSystem {
	return &LightSystem{
		sendGate:      make(chan struct{}, 1),
		revisions:     make(map[Light]uint64),
		lights:        make(map[Light]LightInfo),
		commandClient: commandClient,
	}
}

// Get returns the latest reported or optimistically requested state.
// It returns [ErrLightUnavailable] if the printer hasn't reported this light yet.
func (l *LightSystem) Get(id Light) (LightInfo, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	lt, ok := l.lights[id]
	if !ok {
		return LightInfo{}, ErrLightUnavailable
	}

	return lt, nil
}

// Set updates local state immediately, then sends the command. A failed send
// restores the previous state unless a printer report has superseded it.
// Later reports remain authoritative.
func (l *LightSystem) Set(ctx context.Context, id Light, mode LightMode) error {
	ctx, cancel := withDefaultOpTimeout(ctx)
	defer cancel()
	select {
	case l.sendGate <- struct{}{}:
		defer func() { <-l.sendGate }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	switch mode {
	case LightModeOn, LightModeOff, LightModeFlashing:
	default:
		return ErrInvalidLightMode
	}

	l.mu.Lock()
	previous, ok := l.lights[id]
	if !ok {
		l.mu.Unlock()
		return ErrLightUnavailable
	}

	l.revisions[id]++
	revision := l.revisions[id]
	l.lights[id] = LightInfo{Light: id, Mode: mode}
	l.mu.Unlock()

	if err := l.commandClient.Send(ctx, newLightCommand(id, mode, DefaultLightFlashingConfig())); err != nil {
		l.mu.Lock()
		if l.revisions[id] == revision {
			l.lights[id] = previous
		}
		l.mu.Unlock()
		return err
	}
	return nil
}

// apply records a light state reported by the printer. Called by
// [LightDecoder] while holding the printer's decode lock.
func (l *LightSystem) apply(id Light, mode LightMode) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.revisions[id]++
	l.lights[id] = LightInfo{Light: id, Mode: mode}
}

func newLightCommand(light Light, mode LightMode, cfg LightFlashingConfig) *protocol.Command {
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
		p.Lights().apply(Light(rawLight.Node), LightMode(rawLight.Mode))
	}
}
