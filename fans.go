package bambulabs_api

type Fan uint8

const (
	PartCoolingFan Fan = 1
	AuxiliaryFan   Fan = 2
	ChamberFan     Fan = 3
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
