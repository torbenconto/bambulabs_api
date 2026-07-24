package bambulabs_api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	bambulabs_api "github.com/torbenconto/bambulabs_api"
)

func TestFanSystem_Emulator_Fixtures(t *testing.T) {
	type fanCase struct {
		name           string
		fan            bambulabs_api.Fan
		initialPercent int
		targetPercent  int
	}

	tests := []struct {
		name       string
		model      bambulabs_api.Model
		reportFile string
		cases      []fanCase
	}{
		{
			name:       "h2dpro",
			model:      bambulabs_api.ModelH2DPro,
			reportFile: "h2dpro.json",
			cases: []fanCase{
				{"part_cooling", bambulabs_api.PartCoolingFan, 0, 60},
				{"chamber", bambulabs_api.ChamberFan, 0, 80},
			},
		},
		{
			name:       "x2d",
			model:      bambulabs_api.ModelX2D,
			reportFile: "x2d.json",
			cases: []fanCase{
				{"part_cooling", bambulabs_api.PartCoolingFan, 70, 20},
				{"chamber", bambulabs_api.ChamberFan, 80, 100},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, emu := startEmulatedPrinter(t, bambulabs_api.NewPrinter, tc.model, tc.reportFile)

			for _, fc := range tc.cases {
				t.Run(fc.name, func(t *testing.T) {
					got, err := p.Fans.Get(fc.fan)
					require.NoError(t, err)
					require.Equal(t, fc.initialPercent, got.Percent,
						"initial state decoded from fixture did not match expected value")

					require.NoError(t, p.Fans.Set(context.Background(), fc.fan, fc.targetPercent))

					require.Eventually(t, func() bool {
						got, err := p.Fans.Get(fc.fan)
						return err == nil && got.Percent == fc.targetPercent
					}, 2*time.Second, 20*time.Millisecond,
						"printer did not observe updated fan state from emulator")
				})
			}

			t.Run("auxiliary/unavailable", func(t *testing.T) {
				err := p.Fans.Set(context.Background(), bambulabs_api.AuxillaryFan, 50)
				require.ErrorIs(t, err, bambulabs_api.ErrFanUnavailable)
			})

			state := emu.State()
			require.NotNil(t, state.Print)
		})
	}
}

func TestFanSystem_Emulator_AuxFanAvailable(t *testing.T) {
	p, emu := startEmulatedPrinter(t, bambulabs_api.NewPrinter, bambulabs_api.ModelA1, "mock/all_fans.json")

	got, err := p.Fans.Get(bambulabs_api.AuxillaryFan)
	require.NoError(t, err)
	require.Equal(t, 0, got.Percent)

	require.NoError(t, p.Fans.Set(context.Background(), bambulabs_api.AuxillaryFan, 80))

	require.Eventually(t, func() bool {
		got, err := p.Fans.Get(bambulabs_api.AuxillaryFan)
		return err == nil && got.Percent == 80
	}, 2*time.Second, 20*time.Millisecond, "printer did not observe updated fan state from emulator")

	state := emu.State()
	require.NotNil(t, state.Print)
}
