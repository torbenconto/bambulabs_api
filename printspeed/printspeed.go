package printspeed

type PrintSpeed int

const (
	Silent PrintSpeed = iota + 1
	Standard
	Sport
	Ludicrous
)

func (s PrintSpeed) String() string {
	switch s {
	case Silent:
		return "Silent"
	case Standard:
		return "Standard"
	case Sport:
		return "Sport"
	case Ludicrous:
		return "Ludicrous"
	default:
		return "Unknown"
	}
}
