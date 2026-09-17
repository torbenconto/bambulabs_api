package emulator

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestStateReturnsIndependentSnapshot(t *testing.T) {
	e := &Emulator{state: protocol.Report{Print: &protocol.PrintReport{
		HMSErrors: []protocol.HMSErrorReport{
			{Attr: 1, Code: 2},
		},
		LightsReport: []protocol.LightsReport{
			{Node: "chamber_light", Mode: "off"},
		},
		AMS: &protocol.AMSReport{AMS: []protocol.AMSUnitReport{{
			Tray: []protocol.TrayReport{{Cols: []string{"FF0000"}}},
		}}},
	}}}

	snapshot := e.State()
	snapshot.Print.HMSErrors[0].Code = 99
	snapshot.Print.LightsReport[0].Mode = "on"

	snapshot.Print.AMS.AMS[0].Tray[0].Cols[0] = "000000"

	got := e.State()

	require.Equal(t, uint32(2), got.Print.HMSErrors[0].Code)
	require.Equal(t, "off", got.Print.LightsReport[0].Mode)
	require.Equal(t, "FF0000", got.Print.AMS.AMS[0].Tray[0].Cols[0])

	e.applyLedCtrl(map[string]any{"led_node": "chamber_light", "led_mode": "flashing"})
	require.Equal(t, "off", got.Print.LightsReport[0].Mode)
}

func TestStateConcurrentWithCommands(t *testing.T) {
	e := &Emulator{state: protocol.Report{Print: &protocol.PrintReport{
		LightsReport: []protocol.LightsReport{{Node: "chamber_light", Mode: "off"}},
	}}}

	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			e.applyLedCtrl(map[string]any{"led_node": "chamber_light", "led_mode": "on"})
		}
	})
	wg.Go(func() {
		for range 100 {
			snapshot := e.State()
			snapshot.Print.LightsReport[0].Mode = "off"
		}
	})
	wg.Wait()
	require.Equal(t, "on", e.State().Print.LightsReport[0].Mode)
}
