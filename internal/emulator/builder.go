package emulator

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
)

type MessageBuilder struct {
	msg *mqtt.Message
}

func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		msg: &mqtt.Message{},
	}
}

func (m *MessageBuilder) SetCapability(capability bambulabs_api.Capability) {
	p := &m.msg.Print

	if bambulabs_api.HasCapability(capability, bambulabs_api.CapabilityAms) {
		p.AmsStatus = 0          // OK
		p.AmsRfidStatus = 6      // OK?
		p.Ams.AmsExistBits = "1" // OK
		p.Ams.InsertFlag = true
		p.Ams.PowerOnFlag = false
		p.Ams.TrayExistBits = "e"
		p.Ams.TrayIsBblBits = "e"
		p.Ams.TrayNow = "255"
		p.Ams.TrayPre = "255"
		p.Ams.TrayReadingBits = "0"
		p.Ams.TrayReadDoneBits = "e"
		p.Ams.TrayTar = "255"
		p.Ams.Version = 4
		p.Ams.Ams = append(p.Ams.Ams, mqtt.AMSUnit{
			ID:       "0",
			Humidity: strconv.Itoa((randInt(0, 5))),              // "0" -> "4", humidity value
			Temp:     fmt.Sprintf("%.1f", randFloat(20.0, 24.0)), // room temp in deg C
		})

		for _, id := range []string{"0", "1", "2", "3"} {
			p.Ams.Ams[0].Tray = append(p.Ams.Ams[0].Tray, randTray(id, .25)) // only support 1 ams right now, 25% chance to be empty
		}

		// set vt_tray (external spool), larger chance to be empty
		p.VtTray = randTray("254", .75)
	}
}

func (m *MessageBuilder) SetGcodeState(state bambulabs_api.GcodeState) {
	p := &m.msg.Print
	switch state {
	case bambulabs_api.RUNNING:
		m.resetPrintState()
		totalLayers := randInt(100, 400)
		currentLayer := randInt(1, totalLayers)

		p.LayerNum = currentLayer
		p.TotalLayerNum = totalLayers
		p.McPercent = int(float64(currentLayer) / float64(totalLayers) * 100)

		p.McPrintStage = strconv.Itoa(randInt(2, 5))
		p.McPrintSubStage = randInt(0, 3)
		p.McRemainingTime = randInt(300, 7200)
		p.McPrintErrorCode = "0"

		p.NozzleTargetTemper = randFloat(200, 230)
		p.NozzleTemper = p.NozzleTargetTemper - randFloat(0, 5)

		bed := randFloat(50.0, 60.0)
		p.BedTemper = bed
		p.BedTargetTemper = bed

		p.PrintRealAction = 1
		p.PrintGcodeAction = 1
		p.PrintError = 0
		p.PrintType = "local"

		p.GcodeState = string(bambulabs_api.RUNNING)
		p.GcodeFile = "example.gcode"
		p.GcodeFilePreparePercent = "100"
		p.GcodeStartTime = strconv.FormatInt(time.Now().Add(-time.Duration(randInt(60, 3600))*time.Second).Unix(), 10)

		p.SubtaskName = "test"
		p.SubtaskID = "test"
		p.TaskID = strconv.Itoa(randInt(100000, 999999))
		p.ProjectID = strconv.Itoa(randInt(100000, 999999))
		p.ProfileID = strconv.Itoa(randInt(100000, 999999))
		p.QueueNumber = 0

	case bambulabs_api.IDLE:
	case bambulabs_api.UNKNOWN:
	default:
		m.resetPrintState()

	}
}

func (m *MessageBuilder) resetPrintState() {
	p := &m.msg.Print

	p.LayerNum = 0
	p.TotalLayerNum = 0
	p.McPercent = 0
	p.McPrintErrorCode = "0"
	p.McPrintStage = "1"
	p.McPrintSubStage = 0
	p.McRemainingTime = 0
	p.NozzleTemper = 25.0
	p.NozzleTargetTemper = 25.0
	p.PrintError = 0
	p.PrintGcodeAction = 0
	p.PrintRealAction = 0
	p.PrintType = ""
	p.ProfileID = ""
	p.ProjectID = ""
	p.QueueNumber = 0
	p.TaskID = ""
	p.SubtaskID = ""
	p.SubtaskName = ""
	p.GcodeState = string(bambulabs_api.IDLE)
	p.GcodeFile = ""
	p.GcodeFilePreparePercent = "0"
	p.GcodeStartTime = "0"
}

func randTray(id string, emptyChance float32) mqtt.Tray {
	if rand.Float32() < emptyChance {
		return mqtt.Tray{
			ID: id,
		}
	}

	trayColor := randColor()

	return mqtt.Tray{
		ID: id,
		// metadata for the filament, we'll leave it empty for now
		BedTemp:       "0",
		BedTempType:   "0",
		DryingTemp:    "0",
		DryingTime:    "0",
		TrayDiameter:  "0.00",
		Remain:        0,
		TagUID:        "0000000000000000",                 // 16
		XcamInfo:      "000000000000000000000000",         // 24
		TrayUUID:      "00000000000000000000000000000000", // 32
		TraySubBrands: "",
		TrayIDName:    "",
		TrayWeight:    "0",

		// sane defaults, always defined in the mqtt message
		NozzleTempMax: "240",
		NozzleTempMin: "190",
		TrayType:      "PLA",   // PLA for now
		TrayInfoIdx:   "GFA00", // some sort of ident

		TrayColor: trayColor,
		Cols: []string{
			trayColor,
		},
	}
}

// random RGBA color
func randColor() string {
	bytes := make([]byte, 3)
	crand.Read(bytes) // "It never returns an error, and always fills b entirely."

	bytes = append(bytes, 0xFF)
	return hex.EncodeToString(bytes)
}

func randInt(min, max int) int {
	return rand.IntN(max-min) + min
}

func randFloat(min, max float64) float64 {
	return (rand.Float64() * (max - min)) + min
}
