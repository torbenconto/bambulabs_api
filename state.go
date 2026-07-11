package bambulabs_api

// GcodeState is an enum representing the current print state as dictated by printer.
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
