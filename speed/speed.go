package speed

type Speed int

const (
	Silent Speed = iota + 1
	Standard
	Sport
	Ludicrous
)

func (s Speed) String() string {
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
