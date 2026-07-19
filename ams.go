package bambulabs_api

import (
	"image/color"
	"strconv"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

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
	vt  Tray // "vitrual tray", external spool outside of ams
}

func (a AMSSystem) Units() []AMS {
	return a.ams
}

func (a AMSSystem) ExternalTray() Tray {
	return a.vt
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
	ID            int
	HumidityLevel int

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
	return tray != nil && tray.Filament.RemainingPercent > 0
}

type Tray struct {
	Slot int

	Filament FilamentInfo

	RFID RFIDInfo

	TemperatureInfo TemperatureRequirements
}

type FilamentInfo struct {
	RemainingPercent int
	Material         string
	Diameter         float32
	Color            color.RGBA

	// For multi-color filament
	Colors []color.RGBA
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

type AMSDecoder struct {
	model         Model
	commandClient CommandClient
}

func NewAMSDecoder(model Model, commandClient CommandClient) *AMSDecoder {
	return &AMSDecoder{
		model:         model,
		commandClient: commandClient,
	}
}

func (a *AMSDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil || report.Print.AMS == nil {
		return
	}

	rawAMSUnits := report.Print.AMS.AMS
	if len(rawAMSUnits) == 0 /* [[unlikely]] */ {
		return
	}

	p.cap.Add(CapabilityAMS)

	decodedUnits := make([]AMS, 0, len(rawAMSUnits))
	for _, rawUnit := range rawAMSUnits {
		decodedUnit := AMS{ // TODO: add humitidity and drying stuff
			ID:            parseInt(rawUnit.ID),
			Model:         a.decodeAMSModel(rawUnit.Info), // info is "" if empty
			Trays:         make([]Tray, 0, len(rawUnit.Tray)),
			HumidityLevel: parseInt(rawUnit.Humidity),
		}

		for _, rawTray := range rawUnit.Tray {
			decodedUnit.Trays = append(decodedUnit.Trays, a.decodeTray(&rawTray))
		}

		decodedUnits = append(decodedUnits, decodedUnit)
	}

	p.AMS.ams = decodedUnits
	p.AMS.vt = a.decodeTray(report.Print.VtTray)
}

func (a *AMSDecoder) decodeAMSModel(info string) AMSModel {
	if info == "" {
		return a.defaultAMSModel()
	}

	return decodeAMSInfo(info).Model
}

func (a *AMSDecoder) defaultAMSModel() AMSModel {
	switch a.model {
	case ModelA1, ModelA1Mini:
		return AMSModelLite
	default:
		return AMSModelBase
	}
}

type amsInfo struct {
	Model AMSModel
	// BoundExtruders   []ExtruderID
	SwitcherPosition uint8
}

func decodeAMSInfo(info string) amsInfo {
	var result amsInfo

	raw, err := strconv.ParseUint(info, 16, 64)
	if err != nil {
		// result.BoundExtruders = []ExtruderID{MainExtruder}
		return result
	}

	result.Model = AMSModel(
		getFlagBits(raw, 0, 4),
	)

	// extruderID := getFlagBits(raw, 8, 4)

	// if extruderID == 0xE {
	// 	if hasFilamentSwitch {
	// 		bindSwitch := getFlagBits(raw, 24, 4)

	// 		if bindSwitch == 0 || bindSwitch == 1 {
	// 			result.BoundExtruders = []ExtruderID{
	// 				MainExtruder,
	// 				DeputyExtruder,
	// 			}
	// 		}

	// 		if bindSwitch == 0 {
	// 			result.SwitcherPosition = 0 // POS_IN_B
	// 		} else {
	// 			result.SwitcherPosition = 1 // POS_IN_A
	// 		}
	// 	} else {
	// 		result.BoundExtruders = []ExtruderID{}
	// 	}
	// } else {
	// 	result.BoundExtruders = []ExtruderID{ExtruderID(extruderID)}
	// }

	return result
}

func getFlagBits(value uint64, offset uint, size uint) uint64 {
	mask := uint64((1 << size) - 1)
	return (value >> offset) & mask
}

func (a *AMSDecoder) decodeTray(raw *protocol.TrayReport) Tray {

	decodedColors := make([]color.RGBA, 0, len(raw.Cols))
	for _, col := range raw.Cols {
		decodedColors = append(decodedColors, decodeColor(col))
	}

	filamentInfo := FilamentInfo{
		RemainingPercent: raw.Remaining,
		Color:            decodeColor(raw.TrayColor),
		Colors:           decodedColors,
		Diameter:         parseFloat32(raw.TrayDiameter),
		Material:         raw.TrayType,
	}

	tray := Tray{
		Slot: parseInt(raw.ID),
		RFID: RFIDInfo{
			UID:  raw.TagUID,
			UUID: raw.TrayUUID,
		},

		TemperatureInfo: TemperatureRequirements{
			MinNozzleTemp: parseInt(raw.NozzleTempMin),
			MaxNozzleTemp: parseInt(raw.NozzleTempMax),
			BedTemp:       parseInt(raw.BedTemp),
		},

		Filament: filamentInfo,
	}

	return tray
}
