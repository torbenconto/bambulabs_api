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
