package bambulabs_api

import "slices"

type Fan int

const (
	PartCoolingFan Fan = iota + 1
	AuxiliaryFan
	ChamberFan
)

func (f Fan) String() string {
	switch f {
	case PartCoolingFan:
		return "Part Cooling Fan"
	case AuxiliaryFan:
		return "Auxiliary Fan"
	case ChamberFan:
		return "Chamber Fan"
	default:
		return "Unknown"
	}
}

var allFans = []Fan{PartCoolingFan, AuxiliaryFan, ChamberFan}

var fansForModel = map[Model][]Fan{
	ModelUnknown: {},

	ModelA1Mini: {PartCoolingFan},
	ModelA1:     {PartCoolingFan},

	ModelP1S: allFans,
	ModelP2S: allFans,

	ModelX1C: allFans,
	ModelX1E: allFans,
	ModelX2D: allFans,

	ModelH2:     allFans,
	ModelH2S:    allFans,
	ModelH2D:    allFans,
	ModelH2DPro: allFans,
	ModelH2C:    allFans,
}

func FansForModel(m Model) []Fan {
	return fansForModel[m]
}

func SupportsFan(m Model, f Fan) bool {
	if !slices.Contains(FansForModel(m), f) {
		return false
	}

	return false
}
