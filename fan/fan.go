package fan

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
