package bambulabs_api

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
