package bambulabs_api

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAMSSystem_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		ams := AMSSystem{ams: []AMS{{ID: 1}, {ID: 3}}}

		require.NotNil(t, ams.Get(1))
		require.NotNil(t, ams.Get(3))
	})

	t.Run("not found", func(t *testing.T) {
		ams := AMSSystem{ams: []AMS{{ID: 1}, {ID: 3}}}

		require.Nil(t, ams.Get(2))
	})

	t.Run("empty system", func(t *testing.T) {
		ams := AMSSystem{}

		require.Nil(t, ams.Get(0))
	})
}

func TestAMS_Tray(t *testing.T) {
	ams := AMS{
		Trays: []Tray{{}, {}},
	}

	require.NotNil(t, ams.Tray(0))
	require.NotNil(t, ams.Tray(1))
	assert.Nil(t, ams.Tray(-1))
	assert.Nil(t, ams.Tray(2))
}

func TestTray_HasFilament(t *testing.T) {
	t.Run("no percent, no material", func(t *testing.T) {
		tray := Tray{Filament: FilamentInfo{RemainingPercent: 0, Material: ""}}
		assert.False(t, tray.HasFilament())
	})

	t.Run("percent only", func(t *testing.T) {
		tray := Tray{Filament: FilamentInfo{RemainingPercent: 100, Material: ""}}
		assert.True(t, tray.HasFilament())
	})

	t.Run("material only, zero percent", func(t *testing.T) {
		tray := Tray{Filament: FilamentInfo{RemainingPercent: 0, Material: "PLA"}}
		assert.True(t, tray.HasFilament())
	})

	t.Run("both set", func(t *testing.T) {
		tray := Tray{Filament: FilamentInfo{RemainingPercent: 100, Material: "PLA"}}
		assert.True(t, tray.HasFilament())
	})
}
func TestAMSDecoder(t *testing.T) {
	t.Run("a1/default AMS-Lite", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "a1.json")

		got := requireAMS(t, p, 0)

		assert.Equal(t,
			expectedA1AMS(),
			*got,
		)
	})

	t.Run("a1/external tray", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "a1.json")

		require.NotNil(t, p.AMS)

		assert.Equal(t,
			expectedA1ExternalTray(),
			p.AMS.vt,
		)
	})

	t.Run("h2dpro/explicit AMS-Pro + AMS-HT", func(t *testing.T) {
		p := newTestPrinter(t, ModelH2DPro, "h2dpro.json")

		got := requireAMS(t, p, 0)

		assert.Equal(t,
			expectedH2DProAMS0(),
			*got,
		)

		got1 := requireAMS(t, p, 1)
		assert.Equal(t,
			expectedH2DProAMS1(),
			*got1,
		)

	})

	t.Run("p1/no AMS", func(t *testing.T) {
		p := newTestPrinter(t, ModelP1P, "p1p_no_ams.json")

		require.NotNil(t, p.AMS) // AmsSystem should still be present even with no ams in data
		require.False(t, p.cap.Has(CapabilityAMS))
		require.Len(t, p.AMS.ams, 0) // dont use .Units(), move that to unit test
	})
}

func requireAMS(t *testing.T, p *printer, id int) *AMS {
	t.Helper()

	require.NotNil(t, p.AMS)
	require.True(t, p.cap.Has(CapabilityAMS))
	require.GreaterOrEqual(t, id, 0)
	require.Less(t, id, len(p.AMS.ams))

	return &p.AMS.ams[id]
}

func expectedH2DProAMS0() AMS {
	return AMS{
		ID:            0,
		Model:         AMSModelPro,
		HumidityLevel: 5,

		Trays: []Tray{
			expectedH2DProAMS0Tray(0),
			expectedH2DProAMS0Tray(1),
			expectedH2DProAMS0Tray(2),
			expectedH2DProAMS0Tray(3),
		},
	}
}

func expectedH2DProAMS1() AMS {
	return AMS{
		ID:            128,
		Model:         AMSModelHighTemp,
		HumidityLevel: 0,
		Trays: []Tray{
			expectedH2DProAMS1Tray0(),
		},
	}
}

func expectedH2DProAMS1Tray0() Tray {
	trayColor := color.RGBA{0xC1, 0x2E, 0x1F, 0xFF}

	return Tray{
		Slot: 0,

		Filament: FilamentInfo{
			RemainingPercent: 87,
			Material:         "PLA",
			Diameter:         1.75,
			Color:            trayColor,
			Colors: []color.RGBA{
				trayColor,
			},
		},

		RFID: RFIDInfo{
			UID:  "B10B8F0F00080100",
			UUID: "28DB4DCAB3C348CC9C29E16D52DABFCC",
		},

		TemperatureInfo: TemperatureRequirements{
			MinNozzleTemp: 190,
			MaxNozzleTemp: 230,
			BedTemp:       0,
		},
	}
}

func expectedH2DProAMS0Tray(slot int) Tray {
	return Tray{
		Slot: slot,
	}
}

func expectedA1AMS() AMS {
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

func expectedA1ExternalTray() Tray {
	trayColor := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	return Tray{
		Slot: 254,

		Filament: FilamentInfo{

			RemainingPercent: 0,

			Color: trayColor,

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
