package bambulabs_api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestFanSetDuringSend(t *testing.T) {
	sendFailure := errors.New("send failed")
	for _, tc := range []struct {
		name    string
		sendErr error
		report  bool
		want    int
	}{
		{"success", nil, false, 60},
		{"failure restores previous state", sendFailure, false, 20},
		{"report overrides successful command", nil, true, 40},
		{"failure preserves newer report", sendFailure, true, 40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var f *FanSystem
			client := commandClientFunc(func(context.Context, *protocol.Command) error {
				got, err := f.Get(PartCoolingFan)
				require.NoError(t, err)
				require.Equal(t, 60, got.Percent, "state must be ready before Send completes")
				// Updates to other fans must not suppress rollback of this fan.
				f.apply(ChamberFan, 80)
				if tc.report {
					f.apply(PartCoolingFan, 40)
				}
				return tc.sendErr
			})
			f = NewFanSystem(client)
			f.apply(PartCoolingFan, 20)
			err := f.Set(t.Context(), PartCoolingFan, 55)
			require.ErrorIs(t, err, tc.sendErr)
			got, err := f.Get(PartCoolingFan)
			require.NoError(t, err)
			require.Equal(t, tc.want, got.Percent)
			other, err := f.Get(ChamberFan)
			require.NoError(t, err)
			require.Equal(t, 80, other.Percent)
		})
	}
}

func TestLightSetDuringSend(t *testing.T) {
	sendFailure := errors.New("send failed")
	for _, tc := range []struct {
		name    string
		sendErr error
		report  bool
		want    LightMode
	}{
		{"success", nil, false, LightModeOn},
		{"failure restores previous state", sendFailure, false, LightModeOff},
		{"report overrides successful command", nil, true, LightModeFlashing},
		{"failure preserves newer report", sendFailure, true, LightModeFlashing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var l *LightSystem
			client := commandClientFunc(func(context.Context, *protocol.Command) error {
				got, err := l.Get(ChamberLight)
				require.NoError(t, err)
				require.Equal(t, LightModeOn, got.Mode)
				l.apply(WorkLight, LightModeOn)
				if tc.report {
					l.apply(ChamberLight, LightModeFlashing)
				}
				return tc.sendErr
			})
			l = NewLightSystem(client)
			l.apply(ChamberLight, LightModeOff)
			err := l.Set(t.Context(), ChamberLight, LightModeOn)
			require.ErrorIs(t, err, tc.sendErr)
			got, err := l.Get(ChamberLight)
			require.NoError(t, err)
			require.Equal(t, tc.want, got.Mode)
			other, err := l.Get(WorkLight)
			require.NoError(t, err)
			require.Equal(t, LightModeOn, other.Mode)
		})
	}
}

func TestSetRejectsInvalidOrCanceledRequests(t *testing.T) {
	cc := &capturingCommandClient{}
	fans := NewFanSystem(cc)
	fans.apply(PartCoolingFan, 20)
	lights := NewLightSystem(cc)
	lights.apply(ChamberLight, LightModeOff)
	for _, percent := range []int{-1, 101} {
		require.ErrorIs(t, fans.Set(t.Context(), PartCoolingFan, percent), ErrInvalidFanPercent)
	}
	require.ErrorIs(t, lights.Set(t.Context(), ChamberLight, "invalid"), ErrInvalidLightMode)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, fans.Set(ctx, PartCoolingFan, 60), context.Canceled)
	require.ErrorIs(t, lights.Set(ctx, ChamberLight, LightModeOn), context.Canceled)
	require.Zero(t, cc.count())
	fan, err := fans.Get(PartCoolingFan)
	require.NoError(t, err)
	require.Equal(t, 20, fan.Percent)
	light, err := lights.Get(ChamberLight)
	require.NoError(t, err)
	require.Equal(t, LightModeOff, light.Mode)
}

func TestFanDecoderPartialReports(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "mock/all_fans.json")
	require.NoError(t, p.Fans().Set(t.Context(), PartCoolingFan, 60))
	require.NoError(t, p.Fans().Set(t.Context(), AuxillaryFan, 80))
	applyTestReport(t, p, `{"print":{"hms":[]}}`)
	fan, err := p.Fans().Get(PartCoolingFan)
	require.NoError(t, err)
	require.Equal(t, 60, fan.Percent)
	// Auxiliary deltas need not repeat the capability flag.
	applyTestReport(t, p, `{"print":{"big_fan1_speed":"3","cooling_fan_speed":"0"}}`)
	fan, err = p.Fans().Get(AuxillaryFan)
	require.NoError(t, err)
	require.Equal(t, 20, fan.Percent)
	fan, err = p.Fans().Get(PartCoolingFan)
	require.NoError(t, err)
	require.Zero(t, fan.Percent)
}

func TestFanDecoderEmptyReportDoesNotInventFans(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "")
	applyTestReport(t, p, `{"print":{"hms":[]}}`)
	for _, id := range []Fan{PartCoolingFan, ChamberFan, AuxillaryFan} {
		_, err := p.Fans().Get(id)
		require.ErrorIs(t, err, ErrFanUnavailable)
	}
}

func TestSetCancellationWhileSendIsPending(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(CommandClient) func(context.Context) error
	}{
		{"fan", func(client CommandClient) func(context.Context) error {
			f := NewFanSystem(client)
			f.apply(PartCoolingFan, 20)
			return func(ctx context.Context) error { return f.Set(ctx, PartCoolingFan, 60) }
		}},
		{"light", func(client CommandClient) func(context.Context) error {
			l := NewLightSystem(client)
			l.apply(ChamberLight, LightModeOff)
			return func(ctx context.Context) error { return l.Set(ctx, ChamberLight, LightModeOn) }
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{}, 1)
			client := commandClientFunc(func(ctx context.Context, _ *protocol.Command) error {
				started <- struct{}{}
				<-ctx.Done()
				return ctx.Err()
			})
			set := tc.setup(client)
			firstCtx, cancelFirst := context.WithCancel(t.Context())
			firstDone := make(chan error, 1)
			go func() { firstDone <- set(firstCtx) }()
			t.Cleanup(func() {
				cancelFirst()
				<-firstDone
			})
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("first send did not start")
			}
			canceledCtx, cancel := context.WithCancel(t.Context())
			cancel()
			secondDone := make(chan error, 1)
			go func() { secondDone <- set(canceledCtx) }()
			select {
			case err := <-secondDone:
				require.ErrorIs(t, err, context.Canceled)
			case <-time.After(time.Second):
				t.Fatal("canceled call blocked behind an in-flight send")
			}
		})
	}
}
