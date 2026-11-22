package bambulabs_api_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
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

func TestLightSystemFlashingEmulator(t *testing.T) {
	p, emu := startEmulatedPrinter(t, bambulabs_api.ModelA1, "a1.json")
	emu.SetAutoReport(false)

	observer := paho.NewClient(paho.NewClientOptions().
		AddBroker(fmt.Sprintf("mqtts://127.0.0.1:%d", emu.Port())).
		SetClientID("light-command-observer").
		SetUsername("bblp").
		SetPassword("test").
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true})) // emulator uses a self-signed certificate
	t.Cleanup(func() { observer.Disconnect(250) })
	token := observer.Connect()
	require.True(t, token.WaitTimeout(2*time.Second), "observer did not connect")
	require.NoError(t, token.Error())

	commands := make(chan []byte, 1)
	token = observer.Subscribe("device/EMULATOR0001/request", 0, func(_ paho.Client, message paho.Message) {
		commands <- append([]byte(nil), message.Payload()...)
	})
	require.True(t, token.WaitTimeout(2*time.Second), "observer did not subscribe")
	require.NoError(t, token.Error())

	require.NoError(t, p.Lights().SetFlashing(t.Context(), bambulabs_api.ChamberLight, bambulabs_api.LightFlashingConfig{
		OnTime:       1500 * time.Millisecond,
		OffTime:      250 * time.Millisecond,
		LoopTimes:    3,
		IntervalTime: 2 * time.Second,
	}))

	select {
	case payload := <-commands:
		var envelope map[string]map[string]any
		require.NoError(t, json.Unmarshal(payload, &envelope))
		require.Subset(t, envelope["system"], map[string]any{
			"command":       "ledctrl",
			"led_node":      "chamber_light",
			"led_mode":      "flashing",
			"led_on_time":   float64(1500),
			"led_off_time":  float64(250),
			"loop_times":    float64(3),
			"interval_time": float64(2000),
		})
	case <-time.After(2 * time.Second):
		t.Fatal("custom light command was not published")
	}

	require.Eventually(t, func() bool {
		for _, light := range emu.State().Print.LightsReport {
			if light.Node == string(bambulabs_api.ChamberLight) {
				return light.Mode == string(bambulabs_api.LightModeFlashing)
			}
		}
		return false
	}, 2*time.Second, 20*time.Millisecond)
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
