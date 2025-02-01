package light

type Light string

const (
	ChamberLight Light = "chamber_light"
	PartLight    Light = "part_light"
)

func (l Light) String() string {
	switch l {
	case ChamberLight:
		return "Chamber light"
	case PartLight:
		return "Part light"
	default:
		return "Unknown"
	}
}

type Mode string

const (
	Off      Mode = "off"
	On       Mode = "on"
	Flashing Mode = "flashing"
)

func (m Mode) String() string {
	switch m {
	case Off:
		return "Off"
	case On:
		return "On"
	case Flashing:
		return "Flashing"
	default:
		return "Unknown"
	}
}
