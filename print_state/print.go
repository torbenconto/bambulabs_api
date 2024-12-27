package print_state

type PrintState int

const (
	PRINTING PrintState = iota
	AUTO_BED_LEVELING
	HEATBED_PREHEATING
	SWEEPING_XY_MECH_MODE
	CHANGING_FILAMENT
	M400_PAUSE
	PAUSED_FILAMENT_RUNOUT
	HEATING_HOTEND
	CALIBRATING_EXTRUSION
	SCANNING_BED_SURFACE
	INSPECTING_FIRST_LAYER
	IDENTIFYING_BUILD_PLATE_TYPE
	CALIBRATING_MICRO_LIDAR
	HOMING_TOOLHEAD
	CLEANING_NOZZLE_TIP
	CHECKING_EXTRUDER_TEMPERATURE
	PAUSED_USER
	PAUSED_FRONT_COVER_FALLING
	CALIBRATING_LIDAR
	CALIBRATING_EXTRUSION_FLOW
	PAUSED_NOZZLE_TEMPERATURE_MALFUNCTION
	PAUSED_HEAT_BED_TEMPERATURE_MALFUNCTION
	FILAMENT_UNLOADING
	PAUSED_SKIPPED_STEP
	FILAMENT_LOADING
	CALIBRATING_MOTOR_NOISE
	PAUSED_AMS_LOST
	PAUSED_LOW_FAN_SPEED_HEAT_BREAK
	PAUSED_CHAMBER_TEMPERATURE_CONTROL_ERROR
	COOLING_CHAMBER
	PAUSED_USER_GCODE
	MOTOR_NOISE_SHOWOFF
	PAUSED_NOZZLE_FILAMENT_COVERED_DETECTED
	PAUSED_CUTTER_ERROR
	PAUSED_FIRST_LAYER_ERROR
	PAUSED_NOZZLE_CLOG
	UNKNOWN PrintState = -1
	IDLE    PrintState = 255
)

// String returns a human-readable description of the printer state.
func (ps PrintState) String() string {
	switch ps {
	case PRINTING:
		return "The printer is currently printing."
	case AUTO_BED_LEVELING:
		return "The printer is performing an automatic bed leveling."
	case HEATBED_PREHEATING:
		return "The printer is preheating the heatbed."
	case SWEEPING_XY_MECH_MODE:
		return "The printer is performing a sweeping XY mechanical mode."
	case CHANGING_FILAMENT:
		return "The printer is changing the filament."
	case M400_PAUSE:
		return "The printer is paused."
	case PAUSED_FILAMENT_RUNOUT:
		return "The printer is paused due to filament runout."
	case HEATING_HOTEND:
		return "The printer is heating the hotend."
	case CALIBRATING_EXTRUSION:
		return "The printer is calibrating the extrusion."
	case SCANNING_BED_SURFACE:
		return "The printer is scanning the bed surface."
	case INSPECTING_FIRST_LAYER:
		return "The printer is inspecting the first layer."
	case IDENTIFYING_BUILD_PLATE_TYPE:
		return "The printer is identifying the build plate type."
	case CALIBRATING_MICRO_LIDAR:
		return "The printer is calibrating the micro LiDAR."
	case HOMING_TOOLHEAD:
		return "The printer is homing the toolhead."
	case CLEANING_NOZZLE_TIP:
		return "The printer is cleaning the nozzle tip."
	case CHECKING_EXTRUDER_TEMPERATURE:
		return "The printer is checking the extruder temperature."
	case PAUSED_USER:
		return "The printer is paused by the user."
	case PAUSED_FRONT_COVER_FALLING:
		return "The printer is paused due to the front cover falling."
	case CALIBRATING_LIDAR:
		return "The printer is calibrating the LiDAR."
	case CALIBRATING_EXTRUSION_FLOW:
		return "The printer is calibrating the extrusion flow."
	case PAUSED_NOZZLE_TEMPERATURE_MALFUNCTION:
		return "The printer is paused due to a nozzle temperature malfunction."
	case PAUSED_HEAT_BED_TEMPERATURE_MALFUNCTION:
		return "The printer is paused due to a heat bed temperature malfunction."
	case FILAMENT_UNLOADING:
		return "The printer is unloading the filament."
	case PAUSED_SKIPPED_STEP:
		return "The printer is paused due to a skipped step."
	case FILAMENT_LOADING:
		return "The printer is loading the filament."
	case CALIBRATING_MOTOR_NOISE:
		return "The printer is calibrating the motor noise."
	case PAUSED_AMS_LOST:
		return "The printer is paused due to an AMS lost."
	case PAUSED_LOW_FAN_SPEED_HEAT_BREAK:
		return "The printer is paused due to a low fan speed heat break."
	case PAUSED_CHAMBER_TEMPERATURE_CONTROL_ERROR:
		return "The printer is paused due to a chamber temperature control error."
	case COOLING_CHAMBER:
		return "The printer is cooling the chamber."
	case PAUSED_USER_GCODE:
		return "The printer is paused by the user GCODE."
	case MOTOR_NOISE_SHOWOFF:
		return "The printer is showing off the motor noise."
	case PAUSED_NOZZLE_FILAMENT_COVERED_DETECTED:
		return "The printer is paused due to a nozzle filament covered detected."
	case PAUSED_CUTTER_ERROR:
		return "The printer is paused due to a cutter error."
	case PAUSED_FIRST_LAYER_ERROR:
		return "The printer is paused due to a first layer error."
	case PAUSED_NOZZLE_CLOG:
		return "The printer is paused due to a nozzle clog."
	case UNKNOWN:
		return "The printer status is unknown."
	case IDLE:
		return "The printer is idle."
	default:
		return "Unknown state."
	}
}

// GetPrinterState returns the state description based on the index.
func GetPrinterState(index int) PrintState {
	if index < 0 || index > 35 {
		return UNKNOWN
	}
	state := PrintState(index)
	return state
}
