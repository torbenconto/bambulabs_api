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

var fansForModel = map[Model][]Fan{
	ModelUnknown: {},

	ModelA1Mini: {PartCoolingFan},
	ModelA1:     {PartCoolingFan},

	ModelP1S: {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelP2S: {PartCoolingFan, AuxiliaryFan, ChamberFan},

	ModelX1C: {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelX1E: {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelX2D: {PartCoolingFan, AuxiliaryFan, ChamberFan},

	ModelH2:     {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelH2S:    {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelH2D:    {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelH2DPro: {PartCoolingFan, AuxiliaryFan, ChamberFan},
	ModelH2C:    {PartCoolingFan, AuxiliaryFan, ChamberFan},
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
