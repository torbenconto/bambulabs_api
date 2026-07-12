package bambulabs_api

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/torbenconto/bambulabs_api/hms"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Decoder struct{} // stateless

func (d *Decoder) Apply(state *State, msg *protocol.Report) error { // mutates state based on protocol contents
	if msg.Print == nil {
		return nil
	}

	d.decodeAMS(state, msg)
	d.decodeHMS(state, msg)
	d.decodeExtruder(state, msg)
	d.decodeNozzles(state, msg)

	return nil
}

func (d *Decoder) decodeNozzles(state *State, msg *protocol.Report) {
	if msg.Print.Device == nil || msg.Print.Device.Nozzle == nil {
		return
	}

	nozzle := msg.Print.Device.Nozzle

	state.Nozzles.isIdle = true
	state.Nozzles.isRefreshing = false

	for _, nozzleInfo := range nozzle.Info {
		n := parseNozzleInfo(&nozzleInfo, false)
		if n != nil {
			state.Nozzles.extruderNozzles[nozzleInfo.ID] = n
		}
	}

	if nozzle.SrcID >= 0 {
		state.Nozzles.replaceNozzleSrc = &nozzle.SrcID
	}
	if nozzle.TarID >= 0 {
		state.Nozzles.replaceNozzleTar = &nozzle.TarID
	}
}

func parseNozzleInfo(info *protocol.NozzleInfo, onRack bool) *Nozzle {
	if info.ID < 0 {
		return nil
	}

	n := &Nozzle{
		ID:            info.ID,
		SerialNumber:  info.SN,
		FilamentID:    info.FilaID,
		FilamentColor: info.ColorM,
		Diameter:      float32(info.Diameter),
		Wear:          float32(info.Wear),
		PrintTime:     info.TM,
		Status:        info.Stat,
		OnRack:        onRack,
	}

	// Parse type from string (e.g., "HS00")
	if info.Type != "" {
		n.Type = parseNozzleType(info.Type)
		n.FlowType = parseNozzleFlowType(info.Type)
	}

	return n
}

func parseNozzleType(typeStr string) NozzleType {
	// Bambu format: "HS00", "HH00" etc.
	// First char: H = Hardened/other, S = Stainless, etc.
	if len(typeStr) < 2 {
		return NozzleTypeUndefined
	}

	switch typeStr[1] {
	case 'S':
		return NozzleTypeStainlessSteel
	case 'H':
		return NozzleTypeHardenedSteel
	case 'C':
		return NozzleTypeTungstenCarbide
	case 'B':
		return NozzleTypeBrass
	case 'E':
		return NozzleTypeE3D
	default:
		return NozzleTypeUndefined
	}
}

func parseNozzleFlowType(typeStr string) NozzleFlowType {
	// Format: "HS00", "HH00", "HU00"
	// Second char: S = Standard, H = High-Flow, U = TPU High-Flow
	if len(typeStr) < 2 {
		return NozzleFlowTypeNone
	}

	switch typeStr[1] {
	case 'S':
		return NozzleFlowTypeStandard
	case 'H':
		return NozzleFlowTypeHighFlow
	case 'U':
		return NozzleFlowTypeTPUHighFlow
	default:
		return NozzleFlowTypeNone
	}
}

func (d *Decoder) decodeExtruder(state *State, msg *protocol.Report) {

	if msg.Print.Device == nil || msg.Print.Device.Extruder == nil {
		// use legacy single-extruder status message
		state.Extruders.numExtruders = 1
		state.Extruders.currID = MainExtruder
		state.Extruders.targetID = MainExtruder
		state.Extruders.loadingExtruderID = MainExtruder
		state.Extruders.busyLoading = false // ? ??

		ext := &Extruder{
			ID: MainExtruder,

			CurrentTemp: float32(msg.Print.NozzleTemper),
			TargetTemp:  float32(msg.Print.NozzleTargetTemper),
		}

		if msg.Print.AMS != nil {
			ext.CurrentSlot = decodeV1AmsSlot(msg.Print.AMS.TrayNow)
			ext.TargetSlot = decodeV1AmsSlot(msg.Print.AMS.TrayTar)
		}

		state.Extruders.extruders = []*Extruder{ext}
		return
	}

	state.cap.Add(CapabilityDualExtruder)

	extruders := make([]*Extruder, 0, len(msg.Print.Device.Extruder.Info))

	for _, extruder := range msg.Print.Device.Extruder.Info {
		infoBits := uint64(extruder.Info)
		tempBits := uint64(extruder.Temp)
		// statBits := uint64(extruder.Stat)

		ext := &Extruder{
			ID:                ExtruderID(extruder.ID),
			HasFilament:       getFlagBits(infoBits, 1, 1) != 0,
			BufferHasFilament: getFlagBits(infoBits, 2, 1) != 0,
			HasNozzle:         getFlagBits(infoBits, 3, 1) != 0,
			CurrentTemp:       float32(getFlagBits(tempBits, 0, 16)), // TODO: check if actually float
			TargetTemp:        float32(getFlagBits(tempBits, 16, 16)),
			NozzleID:          uint8(extruder.Hnow),
		}

		ext.PreviousSlot = decodeV2AmsSlot(extruder.Spre, state.Extruders.numExtruders)
		ext.CurrentSlot = decodeV2AmsSlot(extruder.Snow, state.Extruders.numExtruders)
		ext.TargetSlot = decodeV2AmsSlot(extruder.Star, state.Extruders.numExtruders)

		for _, backupRaw := range extruder.FilamBak {
			if backupMap, ok := backupRaw.(float64); ok {
				// Convert backup bits to tray IDs
				backupInt := uint32(backupMap)
				ext.BackupSlots = append(ext.BackupSlots, backupInt)
			}
		}

		// Filament change step (bits 0:8 of stat field)
		// filamentStep := FilamentStep(getFlagBits(statBits, 0, 8))

		extruders = append(extruders, ext)
	}

	state.Extruders.extruders = extruders
}

func decodeV1AmsSlot(tray string) FilamentSlot {
	if tray == "" {
		return FilamentSlot{}
	}

	trayInt, err := strconv.Atoi(tray)
	if err != nil {
		return FilamentSlot{}
	}

	if trayInt == 255 {
		return FilamentSlot{AMSId: 0, SlotId: 0}
	}
	if trayInt == 254 {
		return FilamentSlot{AMSId: 255, SlotId: 254}
	}
	if trayInt >= 0x80 && trayInt <= 0x87 {
		return FilamentSlot{AMSId: uint8(trayInt), SlotId: 0}
	}

	return FilamentSlot{
		AMSId:  uint8(trayInt >> 2),
		SlotId: uint8(trayInt & 0x3),
	}
}

func decodeV2AmsSlot(slotBits int, numExtruders int) FilamentSlot {
	// V2.0 uses bit-packed 16-bit field:
	// [0:8]   = slot_id
	// [8:8]   = ams_id
	// Special: 0xffff = empty slot (only for single extruder)

	if slotBits == 0xffff {
		return FilamentSlot{} // Empty
	}

	slotID := uint8(slotBits & 0xFF)
	amsID := uint8((slotBits >> 8) & 0xFF)

	return FilamentSlot{
		AMSId:  amsID,
		SlotId: slotID,
	}
}

func (d *Decoder) decodeHMS(state *State, msg *protocol.Report) {
	if msg.Print.HMSErrors == nil {
		return
	}

	state.Errors = state.Errors[:0]

	for _, err := range msg.Print.HMSErrors {
		state.Errors = append(state.Errors, hms.Error{
			Attribute: err.Attr,
			Code:      err.Code,
		})
	}
}

func (d *Decoder) decodeAMS(state *State, msg *protocol.Report) {
	if msg.Print.AMS == nil {
		return
	}

	if len(msg.Print.AMS.AMS) > 0 {
		state.cap.Add(CapabilityAMS)
	}

	amsUnits := make([]*AMS, 0, len(msg.Print.AMS.AMS))
	for _, unit := range msg.Print.AMS.AMS {
		amsInfo := decodeAMSInfo(unit.Info, false) // only used for model for now...
		ams := &AMS{
			ID:    parseInt(unit.ID),
			Model: amsInfo.Model,
		}

		for _, tray := range unit.Tray {
			ams.Trays = append(ams.Trays, decodeTray(&tray))
		}

		amsUnits = append(amsUnits, ams)
	}

	state.AMS.ams = amsUnits
}

func decodeAMSInfo(info string, hasFilamentSwitch bool) AMSInfo {
	var result AMSInfo

	raw, err := strconv.ParseUint(info, 16, 64)
	if err != nil {
		// result.BoundExtruders = []uint8{MainExtruder}
		return result
	}

	result.Model = AMSModel(
		getFlagBits(raw, 0, 4),
	)

	extruderID := getFlagBits(raw, 8, 4)

	if extruderID == 0xE {
		if hasFilamentSwitch {
			bindSwitch := getFlagBits(raw, 24, 4)

			if bindSwitch == 0 || bindSwitch == 1 {
				// result.BoundExtruders = []uint8{
				// 	MainExtruder,
				// 	DeputyExtruder,
				// }
			}

			if bindSwitch == 0 {
				result.SwitcherPosition = 0 // POS_IN_B
			} else {
				result.SwitcherPosition = 1 // POS_IN_A
			}
		} else {
			result.BoundExtruders = []uint8{}
		}
	} else {
		result.BoundExtruders = []uint8{uint8(extruderID)}
	}

	return result
}

func getFlagBits(value uint64, offset uint, size uint) uint64 {
	mask := uint64((1 << size) - 1)
	return (value >> offset) & mask
}

func decodeTray(raw *protocol.TrayReport) Tray {
	tray := Tray{
		Color: decodeColor(raw.TrayColor),

		Diameter: parseFloat32(raw.TrayDiameter),

		RFIDUID: raw.TagUID,
		UUID:    raw.TrayUUID,

		BedTemp: parseInt(raw.BedTemp),

		MinNozzleTemp: parseInt(raw.NozzleTempMin),
		MaxNozzleTemp: parseInt(raw.NozzleTempMax),
	}

	for _, col := range raw.Cols {
		tray.Colors = append(tray.Colors, decodeColor(col))
	}

	remain := RemainingFilament{
		Percent: (parseInt(raw.TrayWeight) * raw.Remaining) / 100,
		Grams:   nil,
	}

	if raw.RemainingGrams != nil {
		remain.Grams = raw.RemainingGrams
	}

	tray.Remaining = remain

	return tray
}

func decodeColor(raw string) color.RGBA {
	defaultColor := color.RGBA{255, 255, 255, 255}

	if raw == "" {
		return defaultColor
	}

	clr, ok := parseColor(raw)
	if ok {
		return clr
	}

	return defaultColor
}

func parseColor(s string) (color.RGBA, bool) {
	s = strings.TrimPrefix(s, "#")

	switch len(s) {
	case 6:
		r, err1 := strconv.ParseUint(s[0:2], 16, 8)
		g, err2 := strconv.ParseUint(s[2:4], 16, 8)
		b, err3 := strconv.ParseUint(s[4:6], 16, 8)
		if err1 != nil || err2 != nil || err3 != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, true

	case 8:
		r, err1 := strconv.ParseUint(s[0:2], 16, 8)
		g, err2 := strconv.ParseUint(s[2:4], 16, 8)
		b, err3 := strconv.ParseUint(s[4:6], 16, 8)
		a, err4 := strconv.ParseUint(s[6:8], 16, 8)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return color.RGBA{}, false
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, true
	}

	return color.RGBA{}, false
}

func parseFloat32(raw string) float32 {
	conv, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return 0.0
	}

	return float32(conv)
}

func parseInt(raw string) int {
	conv, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0
	}

	return int(conv)
}
