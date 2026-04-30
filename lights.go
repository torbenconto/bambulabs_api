package bambulabs_api

import "slices"

type Light string

const (
	ChamberLight Light = "chamber_light"
	WorkLight    Light = "work_light" // Defined but seemingly unused
)

type LightMode string

const (
	LightOn       LightMode = "on"
	LightOff      LightMode = "off"
	LightFlashing LightMode = "flashing"
)

// I know what you're thinking but hear me out, the open-air printers use ChamberLight for their WorkLight which makes no sense
// WorkLight is defined but not used so it'll likely change in the future
var lightsForModel = map[Model][]Light{
	ModelUnknown: {},
	ModelA1Mini:  {ChamberLight},
	ModelA1:      {ChamberLight},
	ModelP1S:     {ChamberLight},
	ModelP2S:     {ChamberLight},
	ModelX1C:     {ChamberLight},
	ModelX1E:     {ChamberLight},
	ModelX2D:     {ChamberLight},
	ModelH2:      {ChamberLight},
	ModelH2S:     {ChamberLight},
	ModelH2D:     {ChamberLight},
	ModelH2DPro:  {ChamberLight},
	ModelH2C:     {ChamberLight},
}

func LightsForModel(m Model) []Light {
	return lightsForModel[m]
}

func SupportsLight(m Model, l Light) bool {
	if !slices.Contains(LightsForModel(m), l) {
		return false
	}

	return true
}
