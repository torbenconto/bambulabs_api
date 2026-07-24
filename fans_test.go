package bambulabs_api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestFanSystem_Get(t *testing.T) {
	t.Run("returns ErrFanUnavailable for a fan the printer never reported", func(t *testing.T) {
		f := NewFanSystem(fakeCommandClient{})

		_, err := f.Get(ChamberFan)
		require.ErrorIs(t, err, ErrFanUnavailable)
	})

	t.Run("returns the last applied state", func(t *testing.T) {
		f := NewFanSystem(fakeCommandClient{})
		f.apply(ChamberFan, 60)

		got, err := f.Get(ChamberFan)
		require.NoError(t, err)
		assert.Equal(t, FanInfo{Fan: ChamberFan, Percent: 60}, got)
	})

	t.Run("apply overwrites the previous value for the same fan", func(t *testing.T) {
		f := NewFanSystem(fakeCommandClient{})
		f.apply(ChamberFan, 60)
		f.apply(ChamberFan, 100)

		got, err := f.Get(ChamberFan)
		require.NoError(t, err)
		assert.Equal(t, 100, got.Percent)
	})

	t.Run("tracks multiple fans independently", func(t *testing.T) {
		f := NewFanSystem(fakeCommandClient{})
		f.apply(PartCoolingFan, 20)
		f.apply(ChamberFan, 80)

		part, err := f.Get(PartCoolingFan)
		require.NoError(t, err)
		assert.Equal(t, 20, part.Percent)

		chamber, err := f.Get(ChamberFan)
		require.NoError(t, err)
		assert.Equal(t, 80, chamber.Percent)
	})
}

func TestFanSystem_Set(t *testing.T) {
	t.Run("returns ErrFanUnavailable without sending a command, for a fan not yet reported", func(t *testing.T) {
		cc := &capturingCommandClient{}
		f := NewFanSystem(cc)

		err := f.Set(context.Background(), ChamberFan, 50)
		require.ErrorIs(t, err, ErrFanUnavailable)
		assert.Empty(t, cc.commands, "Set should not send a command for an unavailable fan")
	})

	t.Run("sends the correct M106 gcode for a known fan", func(t *testing.T) {
		cc := &capturingCommandClient{}
		f := NewFanSystem(cc)
		f.apply(ChamberFan, 0) // must be known before Set will act

		require.NoError(t, f.Set(context.Background(), ChamberFan, 50))

		got := cc.last(t)
		print, ok := got["print"].(map[string]any)
		require.True(t, ok, "expected a print-type command, got %#v", got)

		assert.Equal(t, "gcode_line", print["command"])
		assert.Equal(t, "M106 P3 S128\n", print["param"]) // 50% -> ceil(255*50/100) = 128
	})

	t.Run("uses each fan's own gcode P-index", func(t *testing.T) {
		cases := []struct {
			fan   Fan
			pcode string
		}{
			{PartCoolingFan, "P1"},
			{AuxillaryFan, "P2"},
			{ChamberFan, "P3"},
			{SecondaryAuxillaryFan, "P10"},
		}

		for _, tc := range cases {
			cc := &capturingCommandClient{}
			f := NewFanSystem(cc)
			f.apply(tc.fan, 0)

			require.NoError(t, f.Set(context.Background(), tc.fan, 100))

			got := cc.last(t)
			print := got["print"].(map[string]any)
			assert.Contains(t, print["param"], tc.pcode+" S255", "fan %v", tc.fan)
		}
	})

	t.Run("does not itself update local state -- only a decoded report does", func(t *testing.T) {
		cc := &capturingCommandClient{}
		f := NewFanSystem(cc)
		f.apply(ChamberFan, 0)

		require.NoError(t, f.Set(context.Background(), ChamberFan, 100))

		got, err := f.Get(ChamberFan)
		require.NoError(t, err)
		assert.Equal(t, 0, got.Percent, "Get should reflect the last decoded state, not a pending command")
	})
}

func TestPercentToGCodeSpeed(t *testing.T) {
	tests := []struct {
		percent int
		want    int
	}{
		{0, 0},
		{50, 128}, // ceil(255*50/100)
		{54, 128}, // rounds down to nearest 10 (50) -> same result as 50
		{55, 153}, // halfway case: math.Round rounds away from zero, 5.5 -> 6, so rounds UP to 60
		{100, 255},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, percentToGCodeSpeed(tc.percent), "percent=%d", tc.percent)
	}
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		raw  string
		want int
	}{
		{"", 0},
		{"0", 0},
		{"2", 20}, // raw 2 and 3 both bucket to 20%
		{"3", 20},
		{"15", 100},
		{"not-a-number", 0}, // parseInt failure -> 0, not garbage
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, parsePercent(tc.raw), "raw=%q", tc.raw)
	}
}

func TestFanDecoder_Apply(t *testing.T) {
	t.Run("no-op when report.Print is nil", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")
		NewFanDecoder().Apply(p, &protocol.Report{})

		_, err := p.Fans.Get(PartCoolingFan)
		require.ErrorIs(t, err, ErrFanUnavailable)
	})

	t.Run("decodes part cooling and chamber fans unconditionally", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")

		NewFanDecoder().Apply(p, &protocol.Report{
			Print: &protocol.PrintReport{
				CoolingFanSpeed: "6", // -> 40%
				BigFan2Speed:    "9", // -> 60%
			},
		})

		part, err := p.Fans.Get(PartCoolingFan)
		require.NoError(t, err)
		assert.Equal(t, 40, part.Percent)

		chamber, err := p.Fans.Get(ChamberFan)
		require.NoError(t, err)
		assert.Equal(t, 60, chamber.Percent)
	})

	t.Run("auxiliary fan is neither decoded nor capability-gated when aux_part_fan is false", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")

		NewFanDecoder().Apply(p, &protocol.Report{
			Print: &protocol.PrintReport{
				AuxPartFan:   false,
				BigFan1Speed: "12",
			},
		})

		_, err := p.Fans.Get(AuxillaryFan)
		require.ErrorIs(t, err, ErrFanUnavailable)
		assert.False(t, p.cap.Has(CapabilityAuxFan))
	})

	t.Run("auxiliary fan is decoded and capability-gated when aux_part_fan is true", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")

		NewFanDecoder().Apply(p, &protocol.Report{
			Print: &protocol.PrintReport{
				AuxPartFan:   true,
				BigFan1Speed: "12", // -> 80%
			},
		})

		aux, err := p.Fans.Get(AuxillaryFan)
		require.NoError(t, err)
		assert.Equal(t, 80, aux.Percent)
		assert.True(t, p.cap.Has(CapabilityAuxFan))
	})

	t.Run("a later report updates previously decoded fans", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")
		d := NewFanDecoder()

		d.Apply(p, &protocol.Report{Print: &protocol.PrintReport{CoolingFanSpeed: "3"}})  // 20%
		d.Apply(p, &protocol.Report{Print: &protocol.PrintReport{CoolingFanSpeed: "15"}}) // 100%

		got, err := p.Fans.Get(PartCoolingFan)
		require.NoError(t, err)
		assert.Equal(t, 100, got.Percent)
	})
}
