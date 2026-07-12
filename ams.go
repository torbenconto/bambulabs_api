package bambulabs_api

import "image/color"

// type AMSStatus int

// const (
// 	AMSStatusIdle            AMSStatus = 0x00
// 	AMSStatusFilamentChange            = 0x01
// 	AMSStatusRFIDIdentifying           = 0x02
// 	AMSStatusAssist                    = 0x03
// 	AMSStatusCalibration               = 0x04
// 	AMSStatusColdPull                  = 0x07
// 	AMSStatusSelfCheck                 = 0x10
// 	AMSStatusDebug                     = 0x20
// 	AMSStatusUnknown                   = 0xFF
// )

type AMSModel uint8

const (
	AMSModelExtSpool  AMSModel = 0
	AMSModelBase      AMSModel = 1
	AMSModelLite      AMSModel = 2
	AMSModelPro       AMSModel = 3
	AMSModelHighTemp  AMSModel = 4
	AMSModelLiteMixed AMSModel = 5 // ???
)

type AMSSystem struct {
	ams []*AMS
}

// Represents one AMS, state should contain a []AMS
type AMS struct {
	ID int
	// Serial string

	Model AMSModel
	// Status AMSStatus

	Trays []Tray
}

type AMSInfo struct {
	Model            AMSModel
	BoundExtruders   []uint8
	SwitcherPosition uint8
}

type Tray struct {
	Color  color.RGBA
	Colors []color.RGBA

	Diameter float32

	Remaining RemainingFilament

	RFIDUID string
	UUID    string

	BedTemp int
	// BedTempType seemingly unused

	MinNozzleTemp int
	MaxNozzleTemp int
}

type RemainingFilament struct {
	Percent int
	Grams   *int
}
