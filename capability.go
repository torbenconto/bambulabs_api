package bambulabs_api

import "slices"

type Capability uint8

// We derive light and fan capabilites seperately
const (
	CapabilityCamera Capability = 1 << iota

	CapabilityAmsLite
	CapabilityFullAms
)

// A model that supports full AMS also supports AMS lite.
const CapabilityAnyAms = CapabilityAmsLite | CapabilityFullAms

var allCapabilities = CapabilityAnyAms | CapabilityCamera

type ModelInfo struct {
	Capabilities Capability

	CapableFans   []Fan
	CapableLights []Light
}

var fullyCapable ModelInfo = ModelInfo{
	Capabilities: allCapabilities,
	CapableFans:  allFans,
	CapableLights: []Light{
		ChamberLight,
	},
}

var models = map[Model]ModelInfo{
	ModelUnknown: {
		Capabilities:  0,
		CapableFans:   []Fan{},
		CapableLights: []Light{},
	},

	ModelA1Mini: {
		Capabilities: CapabilityAmsLite | CapabilityCamera,
		CapableFans: []Fan{
			PartCoolingFan,
		},
		CapableLights: []Light{
			ChamberLight,
		},
	},

	ModelA1: {
		Capabilities: CapabilityAmsLite | CapabilityCamera,
		CapableFans: []Fan{
			PartCoolingFan,
		},
		CapableLights: []Light{
			ChamberLight,
		},
	},

	// GUESSED, UNSURE
	ModelA2l: {
		Capabilities: allCapabilities,
		CapableFans: []Fan{
			PartCoolingFan,
		},
		CapableLights: []Light{
			ChamberLight,
		},
	},

	ModelP1S: fullyCapable,

	ModelP2S: fullyCapable,

	ModelX1C: fullyCapable,

	ModelX1E: fullyCapable,

	ModelX2D: fullyCapable,

	ModelH2: fullyCapable,

	ModelH2S: fullyCapable,

	ModelH2D: fullyCapable,

	ModelH2DPro: fullyCapable,

	ModelH2C: fullyCapable,
}

func (c Capability) Has(cap Capability) bool {
	return c&cap != 0
}

func SupportsFan(m Model, f Fan) bool {
	return slices.Contains(models[m].CapableFans, f)
}

func SupportsLight(m Model, l Light) bool {
	return slices.Contains(models[m].CapableLights, l)
}
