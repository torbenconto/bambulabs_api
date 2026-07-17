package bambulabs_api

import "image/color"

type AMSModel uint8

const (
	AMSModelUnknown   AMSModel = 0
	AMSModelBase      AMSModel = 1
	AMSModelLite      AMSModel = 2
	AMSModelPro       AMSModel = 3
	AMSModelHighTemp  AMSModel = 4
	AMSModelLiteMixed AMSModel = 5 // ???
)

func (a AMSModel) String() string {
	switch a {
	case AMSModelUnknown:
		return "Unknown"
	case AMSModelBase:
		return "AMS"
	case AMSModelLite:
		return "AMS-Lite"
	case AMSModelPro:
		return "AMS-Pro"
	case AMSModelHighTemp:
		return "AMS-HT"
	default:
		return ""
	}
}

type AMSSystem struct {
	ams []AMS
}

func (a AMSSystem) Units() []AMS {
	return a.ams
}

func (a AMSSystem) Get(id int) *AMS {
	for i := range a.ams {
		if a.ams[i].ID == id {
			return &a.ams[i]
		}
	}

	return nil
}

type AMS struct {
	ID int

	Model AMSModel

	Trays []Tray
}

func (a AMS) Tray(slot int) *Tray {
	if slot < 0 || slot >= len(a.Trays) {
		return nil
	}

	return &a.Trays[slot]
}

func (a AMS) HasFilament(slot int) bool {
	tray := a.Tray(slot)
	return tray != nil && tray.Filament.Remaining.Percent > 0
}

type Tray struct {
	Slot int

	Color color.RGBA

	// For multi-color filament
	Colors []color.RGBA

	Material string

	Diameter float32

	Filament FilamentInfo

	RFID RFIDInfo

	SuggestedBedTemp int

	TemperatureInfo TemperatureRequirements
}

type FilamentInfo struct {
	Remaining RemainingFilament
}

type RemainingFilament struct {
	Percent int
	Grams   *int
}

type RFIDInfo struct {
	UID  string
	UUID string
}

type TemperatureRequirements struct {
	MinNozzleTemp int
	MaxNozzleTemp int
	BedTemp       int
}
