package state

type GcodeState string

const (
	IDLE    GcodeState = "IDLE"
	PREPARE GcodeState = "PREPARE"
	RUNNING GcodeState = "RUNNING"
	PAUSE   GcodeState = "PAUSE"
	FINISH  GcodeState = "FINISH"
	FAILED  GcodeState = "FAILED"
	UNKNOWN GcodeState = "UNKNOWN"
)

// String returns a human-readable description of the G-code state.
func (gs GcodeState) String() string {
	switch gs {
	case IDLE:
		return "The printer is idle."
	case PREPARE:
		return "The printer is preparing."
	case RUNNING:
		return "The printer is running."
	case PAUSE:
		return "The printer is paused."
	case FINISH:
		return "The printer has finished."
	case FAILED:
		return "The printer has failed."
	case UNKNOWN:
		return "The printer state is unknown."
	default:
		return "Invalid Gcode state."
	}
}
