package bambulabs_api

import (
	_light "github.com/torbenconto/bambulabs_api/light"
	"image/color"
	"reflect"

	"github.com/torbenconto/bambulabs_api/hms"
	"github.com/torbenconto/bambulabs_api/state"
)

type Tray struct {
	ID                int          `json:"id"`                 // ID represents the id of an individual tray
	BedTemperature    float64      `json:"bed_temperature"`    // Bed temperature for the specific filament
	Colors            []color.RGBA `json:"colors"`             // Individual colors in the filament
	DryingTemperature float64      `json:"drying_temperature"` // Temperature for drying the filament (°C)
	DryingTime        int          `json:"drying_time"`        // Drying time (hours)
	NozzleTempMax     float64      `json:"nozzle_temp_max"`    // Maximum nozzle temperature (°C)
	NozzleTempMin     float64      `json:"nozzle_temp_min"`    // Minimum nozzle temperature (°C)
	TrayColor         color.RGBA   `json:"tray_color"`         // Overall filament color
	TrayDiameter      float64      `json:"tray_diameter"`      // Diameter of the filament
	TraySubBrands     string       `json:"tray_sub_brands"`    // Detailed filament type (manual input or Bambu filament)
	TrayType          string       `json:"tray_type"`          // Filament type (e.g., PLA, ABS, PLA-S)
	TrayWeight        int          `json:"tray_weight"`        // Spool weight (grams, in intervals of 250g)
}

type Ams struct {
	Humidity    int     `json:"humidity"`    // 0-5: 0 is dry, 5 is wet
	ID          int     `json:"id"`          // ID of the Ams object
	Temperature float64 `json:"temperature"` // Temperature inside the Ams (°C)
	Trays       []Tray  `json:"trays"`       // List of trays in the Ams
}

type LightReport struct {
	Node _light.Light
	Mode _light.Mode
}

type Data struct {
	Ams                     []Ams            `json:"ams"`                        // List of Ams objects
	AmsExists               bool             `json:"ams_exists"`                 // Whether an Ams is connected
	BedTargetTemperature    float64          `json:"bed_target_temperature"`     // Target bed temperature (°C)
	BedTemperature          float64          `json:"bed_temperature"`            // Current bed temperature (°C)
	AuxiliaryFanSpeed       int              `json:"auxiliary_fan_speed"`        // Speed of the auxiliary fan (0-15)
	ChamberFanSpeed         int              `json:"chamber_fan_speed"`          // Speed of the chamber fan (0-15)
	PartFanSpeed            int              `json:"part_fan_speed"`             // Speed of the cooling fan (0-15)
	HeatbreakFanSpeed       int              `json:"heatbreak_fan_speed"`        // Speed of the heatbreak fan (0-15)
	ChamberTemperature      float64          `json:"chamber_temperature"`        // Current chamber temperature (°C)
	GcodeFile               string           `json:"gcode_file"`                 // Name of the current G-code file
	GcodeFilePreparePercent int              `json:"gcode_file_prepare_percent"` // Print preparation percentage
	GcodeState              state.GcodeState `json:"gcode_state"`                // Current printer state
	HMS                     []hms.Error      `json:"hms"`                        // List of errors (TODO: not fully implemented)
	PrintPercentDone        int              `json:"print_percent_done"`         // Current print completion percentage
	PrintErrorCode          string           `json:"print_error_code"`           // Current print error code
	RemainingPrintTime      int              `json:"remaining_print_time"`       // Estimated remaining print time (minutes)
	NozzleDiameter          string           `json:"nozzle_diameter"`            // Diameter of the nozzle (mm)
	NozzleTargetTemperature float64          `json:"nozzle_target_temperature"`  // Target nozzle temperature (°C)
	NozzleTemperature       float64          `json:"nozzle_temperature"`         // Current nozzle temperature (°C)
	Sdcard                  bool             `json:"sdcard"`                     // Whether an SD card is inserted
	LightReport             []LightReport    `json:"lights_report"`              // Light report
	VtTray                  Tray             `json:"vt_tray"`                    // Built-in tray for use without Ams
	WifiSignal              string           `json:"wifi_signal"`                // Wi-Fi signal strength in dBm
}

// IsEmpty checks if the Data struct is empty using reflection
func (d *Data) IsEmpty() bool {
	dataValue := reflect.ValueOf(d).Elem()
	for i := 0; i < dataValue.NumField(); i++ {
		field := dataValue.Field(i)
		if !field.IsZero() {
			return false
		}
	}
	return true
}
