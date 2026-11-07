package bambulabs_api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestLightSystem_Get(t *testing.T) {
	t.Run("returns ErrLightUnavailable for a light the printer never reported", func(t *testing.T) {
		l := NewLightSystem(&capturingCommandClient{})

		_, err := l.Get(ChamberLight)
		require.ErrorIs(t, err, ErrLightUnavailable)
	})

	t.Run("returns the last applied state", func(t *testing.T) {
		l := NewLightSystem(&capturingCommandClient{})
		l.apply(ChamberLight, LightModeOn)

		got, err := l.Get(ChamberLight)
		require.NoError(t, err)
		assert.Equal(t, LightInfo{Light: ChamberLight, Mode: LightModeOn}, got)
	})

	t.Run("apply overwrites the previous state for the same light", func(t *testing.T) {
		l := NewLightSystem(&capturingCommandClient{})
		l.apply(ChamberLight, LightModeOn)
		l.apply(ChamberLight, LightModeFlashing)

		got, err := l.Get(ChamberLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeFlashing, got.Mode)
	})

	t.Run("tracks multiple lights independently", func(t *testing.T) {
		l := NewLightSystem(&capturingCommandClient{})
		l.apply(ChamberLight, LightModeOn)
		l.apply(WorkLight, LightModeOff)

		chamber, err := l.Get(ChamberLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeOn, chamber.Mode)

		work, err := l.Get(WorkLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeOff, work.Mode)
	})
}

func TestLightSystem_Set(t *testing.T) {
	t.Run("returns ErrLightUnavailable without sending a command, for a light not yet reported", func(t *testing.T) {
		cc := &capturingCommandClient{}
		l := NewLightSystem(cc)

		err := l.Set(context.Background(), ChamberLight, LightModeOn)
		require.ErrorIs(t, err, ErrLightUnavailable)
		assert.Zero(t, cc.count(), "Set should not send a command for an unavailable light")
	})

	t.Run("sends a correctly-shaped ledctrl command for a known light", func(t *testing.T) {
		cc := &capturingCommandClient{}
		l := NewLightSystem(cc)
		l.apply(ChamberLight, LightModeOff) // must be known before Set will act

		require.NoError(t, l.Set(context.Background(), ChamberLight, LightModeOn))

		got := cc.last(t)
		system, ok := got["system"].(map[string]any)
		require.True(t, ok, "expected a system-type command, got %#v", got)

		assert.Equal(t, "ledctrl", system["command"])
		assert.Equal(t, string(ChamberLight), system["led_node"])
		assert.Equal(t, string(LightModeOn), system["led_mode"])

		cfg := DefaultLightFlashingConfig()
		assert.Equal(t, float64(cfg.OnTime.Milliseconds()), system["led_on_time"])
		assert.Equal(t, float64(cfg.OffTime.Milliseconds()), system["led_off_time"])
		assert.Equal(t, float64(cfg.LoopTimes), system["loop_times"])
		assert.Equal(t, float64(cfg.IntervalTime.Milliseconds()), system["interval_time"])
	})

	t.Run("updates local state immediately", func(t *testing.T) {
		cc := &capturingCommandClient{}
		l := NewLightSystem(cc)
		l.apply(ChamberLight, LightModeOff)

		require.NoError(t, l.Set(context.Background(), ChamberLight, LightModeOn))

		got, err := l.Get(ChamberLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeOn, got.Mode)
	})
}

func TestLightSystem_SetFlashing(t *testing.T) {
	cc := &capturingCommandClient{}
	l := NewLightSystem(cc)
	l.apply(WorkLight, LightModeOff)
	cfg := LightFlashingConfig{
		OnTime:       1250 * time.Millisecond,
		OffTime:      750 * time.Millisecond,
		LoopTimes:    4,
		IntervalTime: 2500 * time.Millisecond,
	}

	require.NoError(t, l.SetFlashing(t.Context(), WorkLight, cfg))
	system := cc.last(t)["system"].(map[string]any)
	assert.Subset(t, system, map[string]any{
		"command":       "ledctrl",
		"led_node":      "work_light",
		"led_mode":      "flashing",
		"led_on_time":   float64(1250),
		"led_off_time":  float64(750),
		"loop_times":    float64(4),
		"interval_time": float64(2500),
	})
	light, err := l.Get(WorkLight)
	require.NoError(t, err)
	assert.Equal(t, LightModeFlashing, light.Mode)

	// Custom timing applies only to this call, not subsequent default commands.
	require.NoError(t, l.Set(t.Context(), WorkLight, LightModeFlashing))
	system = cc.last(t)["system"].(map[string]any)
	assert.Subset(t, system, map[string]any{
		"led_mode":      "flashing",
		"led_on_time":   float64(500),
		"led_off_time":  float64(500),
		"loop_times":    float64(1),
		"interval_time": float64(1000),
	})
}

func TestLightSystem_SetFlashingRejectsUnavailableOrCanceledRequests(t *testing.T) {
	cc := &capturingCommandClient{}
	l := NewLightSystem(cc)
	cfg := DefaultLightFlashingConfig()
	require.ErrorIs(t, l.SetFlashing(t.Context(), ChamberLight, cfg), ErrLightUnavailable)

	l.apply(ChamberLight, LightModeOff)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, l.SetFlashing(ctx, ChamberLight, cfg), context.Canceled)
	assert.Zero(t, cc.count())
	light, err := l.Get(ChamberLight)
	require.NoError(t, err)
	assert.Equal(t, LightModeOff, light.Mode)
}

func TestLightDecoder_Apply(t *testing.T) {
	t.Run("no-op when report.Print is nil", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")
		NewLightDecoder().Apply(p, &protocol.Report{})

		_, err := p.Lights().Get(ChamberLight)
		require.ErrorIs(t, err, ErrLightUnavailable)
	})

	t.Run("decodes each reported light", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")

		NewLightDecoder().Apply(p, &protocol.Report{
			Print: &protocol.PrintReport{
				LightsReport: []protocol.LightsReport{
					{Node: string(ChamberLight), Mode: string(LightModeOn)},
					{Node: string(WorkLight), Mode: string(LightModeFlashing)},
				},
			},
		})

		chamber, err := p.Lights().Get(ChamberLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeOn, chamber.Mode)

		work, err := p.Lights().Get(WorkLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeFlashing, work.Mode)
	})

	t.Run("a later report updates previously decoded lights", func(t *testing.T) {
		p := newTestPrinter(t, ModelA1, "")
		d := NewLightDecoder()

		d.Apply(p, &protocol.Report{Print: &protocol.PrintReport{
			LightsReport: []protocol.LightsReport{{Node: string(ChamberLight), Mode: string(LightModeOff)}},
		}})
		d.Apply(p, &protocol.Report{Print: &protocol.PrintReport{
			LightsReport: []protocol.LightsReport{{Node: string(ChamberLight), Mode: string(LightModeOn)}},
		}})

		got, err := p.Lights().Get(ChamberLight)
		require.NoError(t, err)
		assert.Equal(t, LightModeOn, got.Mode)
	})
}
