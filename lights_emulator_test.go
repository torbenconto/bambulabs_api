package bambulabs_api_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/emulator"
)

func TestLightSystemEmulator(t *testing.T) {
	type lightCase struct {
		id          bambulabs_api.Light
		initialMode bambulabs_api.LightMode
		targetMode  bambulabs_api.LightMode
	}

	tests := []struct {
		name       string
		model      bambulabs_api.Model
		reportFile string
		lights     []lightCase
		// missing lights this fixture should report ErrLightNotAvalible for
		unavailable []bambulabs_api.Light
	}{
		{
			name:       "a1",
			model:      bambulabs_api.ModelA1,
			reportFile: "a1.json",
			lights: []lightCase{
				{bambulabs_api.ChamberLight, bambulabs_api.LightModeOff, bambulabs_api.LightModeOn},
			},
			unavailable: []bambulabs_api.Light{
				bambulabs_api.WorkLight,
				bambulabs_api.ChamberLight2,
			},
		},
		{
			name:       "x2d",
			model:      bambulabs_api.ModelX2D,
			reportFile: "x2d.json",
			lights: []lightCase{
				{bambulabs_api.ChamberLight, bambulabs_api.LightModeOn, bambulabs_api.LightModeOff},
				{bambulabs_api.WorkLight, bambulabs_api.LightModeFlashing, bambulabs_api.LightModeOn},
			},
			unavailable: []bambulabs_api.Light{
				bambulabs_api.ChamberLight2,
			},
		},
		{
			name:       "h2dpro",
			model:      bambulabs_api.ModelH2DPro,
			reportFile: "h2dpro.json",
			lights: []lightCase{
				{bambulabs_api.ChamberLight, bambulabs_api.LightModeOn, bambulabs_api.LightModeFlashing},
				{bambulabs_api.WorkLight, bambulabs_api.LightModeFlashing, bambulabs_api.LightModeOff},
				{bambulabs_api.ChamberLight2, bambulabs_api.LightModeOn, bambulabs_api.LightModeOff},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, emu := startEmulatedPrinter(t, tc.model, tc.reportFile)
			emu.SetAutoReport(false)

			for _, lc := range tc.lights {
				t.Run(string(lc.id), func(t *testing.T) {
					assertLightTransition(t, p.Lights(), emu, lc.id, lc.initialMode, lc.targetMode)
				})
			}

			for _, id := range tc.unavailable {
				t.Run(string(id)+"/unavailable", func(t *testing.T) {
					_, err := p.Lights().Get(id)
					require.ErrorIs(t, err, bambulabs_api.ErrLightUnavailable)
				})
			}
		})
	}
}

func assertLightTransition(
	t *testing.T,
	lights *bambulabs_api.LightSystem,
	emu *emulator.Emulator,
	id bambulabs_api.Light,
	initialMode, targetMode bambulabs_api.LightMode,
) {
	t.Helper()

	l, err := lights.Get(id)
	require.NoError(t, err)
	require.Equal(t, initialMode, l.Mode)

	require.NoError(t, lights.Set(context.Background(), id, targetMode))

	l, err = lights.Get(id)
	require.NoError(t, err)
	require.Equal(t, targetMode, l.Mode, "state is ready without a report")

	require.Eventually(t, func() bool {
		state := emu.State()
		if state.Print == nil {
			return false
		}
		for _, reported := range state.Print.LightsReport {
			if reported.Node == string(id) {
				return reported.Mode == string(targetMode)
			}
		}
		return false
	}, 2*time.Second, 20*time.Millisecond, "emulator did not receive light command")
}
