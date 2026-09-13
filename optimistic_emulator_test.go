package bambulabs_api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestOptimisticStateEmulatorReportReconciliation(t *testing.T) {
	p, emu := startEmulatedPrinter(t, bambulabs_api.ModelX2D, "x2d.json")
	emu.SetAutoReport(false)

	require.NoError(t, p.Fans().Set(t.Context(), bambulabs_api.PartCoolingFan, 60))
	require.NoError(t, p.Lights().Set(t.Context(), bambulabs_api.ChamberLight, bambulabs_api.LightModeOff))

	fan, err := p.Fans().Get(bambulabs_api.PartCoolingFan)
	require.NoError(t, err)
	require.Equal(t, 60, fan.Percent)

	light, err := p.Lights().Get(bambulabs_api.ChamberLight)
	require.NoError(t, err)
	require.Equal(t, bambulabs_api.LightModeOff, light.Mode)

	require.NoError(t, emu.PublishReport(protocol.Report{Print: &protocol.PrintReport{
		CoolingFanSpeed: "0",
		LightsReport:    []protocol.LightsReport{{Node: "chamber_light", Mode: "flashing"}},
	}}))

	require.Eventually(t, func() bool {
		fan, fanErr := p.Fans().Get(bambulabs_api.PartCoolingFan)
		light, lightErr := p.Lights().Get(bambulabs_api.ChamberLight)
		return fanErr == nil && lightErr == nil && fan.Percent == 0 && light.Mode == bambulabs_api.LightModeFlashing
	}, 2*time.Second, 10*time.Millisecond)
}
