package bambulabs_api

import "time"

// Light is an enum representing all possible lights and their literal names as used in MQTT communications.
type Light string

const (
	ChamberLight Light = "chamber_light"
	WorkLight    Light = "work_light" // Defined but seemingly unused
)

// LightMode represents supported modes for a given [Light] and their literal string values as used in MQTT commands.
type LightMode string

const (
	LightOn       LightMode = "on"
	LightOff      LightMode = "off"
	LightFlashing LightMode = "flashing"
)

// LightFlashingConfig controls the timing of [LightFlashing].
type LightFlashingConfig struct {
	OnTime       time.Duration
	OffTime      time.Duration
	LoopTimes    int
	IntervalTime time.Duration
}

// DefaultLightFlashingConfig returns the flashing configuration used by [Printer.SetLight].
func DefaultLightFlashingConfig() LightFlashingConfig {
	return LightFlashingConfig{
		OnTime:       500 * time.Millisecond,
		OffTime:      500 * time.Millisecond,
		LoopTimes:    1,
		IntervalTime: time.Second,
	}
}

// I know what you're thinking but hear me out, the open-air printers use ChamberLight for their WorkLight which makes no sense
// WorkLight is defined but not used so it'll likely change in the future
