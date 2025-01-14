package bambulabs_api

import (
	"github.com/torbenconto/bambulabs_api/light"
	"github.com/torbenconto/bambulabs_api/state"
	"image/color"
	"reflect"
)

type Tray struct {
	// ID represents the id of an individual tray
	ID int
	// BedTemperature represents the temperature to which the bed will be heated to for the specific filament loaded in this tray.
	BedTemperature float64
	// Colors represents the individual colors in a filament, this will contain multiple entries for filament of multiple colors
	Colors []color.RGBA
	// DryingTemperature is the temperature at which the filament should be dried in degrees C
	DryingTemperature float64
	// DryingTime is the amount of time this roll of filament should be dried for (hours).
	DryingTime int
	// NozzleTempMax represents the maximum temperature the nozzle should reach whilst printing this filament (deg C).
	NozzleTempMax float64
	// NozzleTempMax represents the maximum temperature the nozzle should reach whilst printing this filament (deg C).
	NozzleTempMin float64
	// TrayColor is the overall color of the filament, for one color filaments this will be the same as Colors[0]
	TrayColor color.RGBA
	// TrayDiameter is the diamter of the filament
	TrayDiameter float64
	// TraySubBrands is a string with the detailed type of filament (only works with bambu filament or if inputted manually)
	TraySubBrands string
	// TrayType is a string which contains the type of filament (eg: PLA ABS PLA-S)
	TrayType string
	// TrayWeight is the estimated weight of the spool in grams (this is typically done in intervals of 250g.
	TrayWeight int
}

type Ams struct {
	// Humidity is a value 0-5 where 0 is dry and 5 is wet
	Humidity int
	// ID is the id of the current Ams object, useful for identifying a single Ams when multiple are present.
	ID int
	// Temperature represents the temperature value inside the Ams in degrees C
	Temperature float64

	Trays []Tray
}

type Data struct {
	// Ams is a list of ams objects
	Ams []Ams

	AmsExists bool

	BedTargetTemperature float64
	BedTemperature       float64
	BigFan1Speed         int
	BigFan2Speed         int
	ChamberTemperature   float64
	CoolingFanSpeed      int

	// GcodeFile is the name of the current gcode file being printed.
	GcodeFile string
	// GcodeFilePreparePercent is the percentage which the current print file is prepared for printing.
	GcodeFilePreparePercent int
	// GcodeState is the current state of the printer.
	GcodeState state.GcodeState

	HeatbreakFanSpeed int
	Hms               []any

	LightsReport []struct {
		// TODO: Make better
		Mode string
		Node light.Light
	}
	PrintPercentDone        int
	PrintErrorCode          string
	RemainingPrintTime      int
	NozzleDiameter          string
	NozzleTargetTemperature float64
	NozzleTemperature       float64
	// Sdcard is a flag which details if an sd card is inserted into the built-in port.
	Sdcard bool

	// VtTray is the built-in tray (for use without Ams)
	VtTray struct {
		// ID represents the id of an individual tray
		ID int
		// BedTemperature represents the temperature to which the bed will be heated to for the specific filament loaded in this tray.
		BedTemperature float64
		// Colors represents the individual colors in a filament, this will contain multiple entries for filament of multiple colors
		Colors []color.RGBA
		// DryingTemperature is the temperature at which the filament should be dried in degrees C
		DryingTemperature float64
		// DryingTime is the amount of time this roll of filament should be dried for (hours).
		DryingTime int
		// NozzleTempMax represents the maximum temperature the nozzle should reach whilst printing this filament (deg C).
		NozzleTempMax float64
		// NozzleTempMax represents the maximum temperature the nozzle should reach whilst printing this filament (deg C).
		NozzleTempMin float64
		// TrayColor is the overall color of the filament, for one color filaments this will be the same as Colors[0]
		TrayColor color.RGBA
		// TrayDiameter is the diamter of the filament
		TrayDiameter float64
		// TraySubBrands is a string with the detailed type of filament (only works with bambu filament or if inputted manually)
		TraySubBrands string
		// TrayType is a string which contains the type of filament (eg: PLA ABS PLA-S)
		TrayType string
		// TrayWeight is the estimated weight of the spool in grams (this is typically done in intervals of 250g.
		TrayWeight int
	}

	// WifiSignal provides the current strength of the wifi signal as a string in dBm
	WifiSignal string
}

// IsEmpty checks if the Data struct is empty using reflection
func (d Data) IsEmpty() bool {
	// Use reflection to iterate over fields and check if they have their zero values
	dataValue := reflect.ValueOf(d).Elem()
	for i := 0; i < dataValue.NumField(); i++ {
		field := dataValue.Field(i)
		// If any field is not the zero value, return false
		if !field.IsZero() {
			return false
		}
	}
	return true
}
