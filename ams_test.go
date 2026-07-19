package bambulabs_api

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAMSDecoder(t *testing.T) {
	t.Run("a1/AMS-Lite", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "a1.json")

		got := requireAMS(t, p, 0)

		assert.Equal(t,
			expectedA1AMSLite(),
			*got,
		)
	})
}

func requireAMS(t *testing.T, p *printer, id int) *AMS {
	t.Helper()

	require.NotNil(t, p.AMS)
	require.True(t, p.cap.Has(CapabilityAMS))

	require.Less(t, id, len(p.AMS.ams))

	return &p.AMS.ams[id]
}

func expectedA1AMSLite() AMS {
	return AMS{
		ID:    0,
		Model: AMSModelLite,

		HumidityLevel: 5,

		Trays: []Tray{
			expectedA1Tray(0, color.RGBA{0x76, 0xD9, 0xF4, 0xFF}),
			expectedA1Tray(1, color.RGBA{0xFF, 0xF1, 0x44, 0xFF}),
			expectedA1Tray(2, color.RGBA{0xFF, 0x80, 0xFF, 0xFF}),
			expectedA1Tray(3, color.RGBA{0x0A, 0xCC, 0x38, 0xFF}),
		},
	}
}

func expectedA1Tray(slot int, trayColor color.RGBA) Tray {
	return Tray{
		Slot: slot,

		Filament: FilamentInfo{

			RemainingPercent: 100,

			Color: trayColor,

			Colors: []color.RGBA{
				trayColor,
			},

			Material: "PLA",

			Diameter: 0.0,
		},

		RFID: RFIDInfo{
			UID:  "0000000000000000",
			UUID: "00000000000000000000000000000000",
		},

		TemperatureInfo: TemperatureRequirements{
			MinNozzleTemp: 190,
			MaxNozzleTemp: 240,
			BedTemp:       0,
		},
	}
}
