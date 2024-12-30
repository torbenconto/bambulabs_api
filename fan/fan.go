package fan

type Fan int

const (
	PartFan Fan = iota + 1
	AuxiliaryFan
	ChamberFan
)

func (f Fan) String() string {
	switch f {
	case PartFan:
		return "Part Fan"
	case AuxiliaryFan:
		return "Auxiliary Fan"
	case ChamberFan:
		return "Chamber Fan"
	default:
		return "Unknown"
	}
}
