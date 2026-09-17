package bambulabs_api

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHMSDecoderReports(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "mock/hms.json")
	initial := p.HMS().Errors()
	require.Len(t, initial, 2)
	require.Equal(t, uint32(0x03000100), initial[0].Attribute)
	require.Equal(t, uint32(0x00010006), initial[0].Code)
	require.Equal(t, "HMS_0300_0100_0001_0006", initial[0].GetCode())
	require.Equal(t, "The heatbed temperature is abnormal; the sensor may have a short circuit.", initial[0].Error())
	require.Equal(t, "HMS_FFFF_FFFF_FFFF_FFFF", initial[1].Error())

	for _, payload := range []string{
		"{}",
		`{"print":{}}`,
		`{"print":{"cooling_fan_speed":"6"}}`,
		`{"print":{"hms":null}}`,
	} {
		t.Run(payload, func(t *testing.T) {
			applyTestReport(t, p, payload)
			require.Equal(t, initial, p.HMS().Errors(), "omitted/null HMS must preserve errors")
		})
	}

	applyTestReport(t, p, `{"print":{"hms":[{"attr":117440512,"code":50364426}]}}`)
	got := p.HMS().Errors()
	require.Len(t, got, 1, "a new HMS list replaces the previous list")
	require.Equal(t, "HMS_0700_0000_0300_800A", got[0].GetCode())

	applyTestReport(t, p, `{"print":{"hms":[]}}`)
	require.Empty(t, p.HMS().Errors(), "an explicitly empty HMS list clears errors")
}

func TestHMSSystemSnapshots(t *testing.T) {
	h := NewHMSSystem()
	require.Empty(t, h.Errors())
	source := []HMSError{{Code: 1, Attribute: 2}}
	h.apply(source)
	source[0].Code = 99
	snapshot := h.Errors()
	require.Equal(t, uint32(1), snapshot[0].Code)
	snapshot[0].Code = 88
	require.Equal(t, uint32(1), h.Errors()[0].Code, "callers cannot mutate system state")
	h.apply([]HMSError{{Code: 3, Attribute: 4}})
	require.Equal(t, uint32(88), snapshot[0].Code, "old snapshots stay independent")
}

func TestHMSSystemConcurrentSnapshots(t *testing.T) {
	h := NewHMSSystem()
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100 {
				h.apply([]HMSError{{Code: 1, Attribute: 2}})
				snapshot := h.Errors()
				if len(snapshot) > 0 {
					snapshot[0].Code = 99
				}
			}
		})
	}
	wg.Wait()
	require.Equal(t, uint32(1), h.Errors()[0].Code)
}

func TestHMSDecoderRejectsInvalidUnsignedFields(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "mock/hms.json")
	initial := p.HMS().Errors()
	for _, payload := range []string{
		`{"print":{"hms":[{"attr":-1,"code":1}]}}`,
		`{"print":{"hms":[{"attr":1,"code":4294967296}]}}`,
	} {
		p.updateState([]byte(payload))
		require.Equal(t, initial, p.HMS().Errors())
	}
}
