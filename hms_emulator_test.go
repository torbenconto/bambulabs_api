package bambulabs_api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestHMSSystemEmulator(t *testing.T) {
	p, emu := startEmulatedPrinter(t, bambulabs_api.ModelA1, "mock/hms.json")
	got := p.HMS().Errors()
	require.Len(t, got, 2)
	require.Equal(t, "HMS_0300_0100_0001_0006", got[0].GetCode())
	require.Equal(t, "The heatbed temperature is abnormal; the sensor may have a short circuit.", got[0].Error())
	require.Equal(t, "HMS_FFFF_FFFF_FFFF_FFFF", got[1].Error())

	require.NoError(t, emu.PublishReport(protocol.Report{Print: &protocol.PrintReport{
		HMSErrors: []protocol.HMSErrorReport{{Attr: 0x07000000, Code: 0x0300800A}},
	}}))
	require.Eventually(t, func() bool {
		errors := p.HMS().Errors()
		return len(errors) == 1 && errors[0].GetCode() == "HMS_0700_0000_0300_800A"
	}, 2*time.Second, 10*time.Millisecond)

	// This must travel as "hms":[]; omitting it would preserve the prior error.
	require.NoError(t, emu.PublishReport(protocol.Report{Print: &protocol.PrintReport{
		HMSErrors: []protocol.HMSErrorReport{},
	}}))
	require.Eventually(t, func() bool {
		return len(p.HMS().Errors()) == 0
	}, 2*time.Second, 10*time.Millisecond)

	emu.PushUpdate()
	require.Eventually(t, func() bool {
		errors := p.HMS().Errors()
		return len(errors) == 2 && errors[0].GetCode() == "HMS_0300_0100_0001_0006"
	}, 2*time.Second, 10*time.Millisecond)
}

func TestHMSSystemEmulatorInitiallyEmpty(t *testing.T) {
	p, _ := startEmulatedPrinter(t, bambulabs_api.ModelA1, "a1.json")
	require.Empty(t, p.HMS().Errors())
}
