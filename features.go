package bambulabs_api

type Fan int

const (
	PartFan Fan = iota + 1
	AuxillaryFan
	ChamberFan
)

func (f Fan) String() string {
	switch f {
	case PartFan:
		return "Part Fan"
	case AuxillaryFan:
		return "Auxillary Fan"
	case ChamberFan:
		return "Chamber Fan"
	default:
		return "Unknown"
	}
}

type Nozzle string

const (
	StainlessSteel Nozzle = "stainless_steel"
	HardenedSteel  Nozzle = "hardened_steel"
)

func (n Nozzle) String() string {
	switch n {
	case StainlessSteel:
		return "Stainless steel"
	case HardenedSteel:
		return "Hardened steel"
	default:
		return "Unknown"
	}
}

type Speed int

const (
	Silent Speed = iota + 1
	Standard
	Sport
	Ludicrous
)
