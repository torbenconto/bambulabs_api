package bambulabs_api

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestWithDefaultOpTimeout(t *testing.T) {
	t.Run("adds default deadline", func(t *testing.T) {
		start := time.Now()
		ctx, cancel := withDefaultOpTimeout(context.Background())
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected a default deadline")
		}
		if got := deadline.Sub(start); got < defaultOpTimeout-time.Second || got > defaultOpTimeout+time.Second {
			t.Fatalf("default deadline = %v, want approximately %v", got, defaultOpTimeout)
		}
	})

	t.Run("preserves caller deadline longer than default", func(t *testing.T) {
		parentDeadline := time.Now().Add(defaultOpTimeout * 2)
		parent, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
		defer parentCancel()

		ctx, cancel := withDefaultOpTimeout(parent)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected the caller deadline")
		}
		if !deadline.Equal(parentDeadline) {
			t.Fatalf("deadline = %v, want %v", deadline, parentDeadline)
		}
	})

	t.Run("preserves caller deadline shorter than default", func(t *testing.T) {
		parentDeadline := time.Now().Add(time.Second)
		parent, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
		defer parentCancel()

		ctx, cancel := withDefaultOpTimeout(parent)
		defer cancel()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected the caller deadline")
		}
		if !deadline.Equal(parentDeadline) {
			t.Fatalf("deadline = %v, want %v", deadline, parentDeadline)
		}
	})
}

func TestNewLightCommand(t *testing.T) {
	cfg := LightFlashingConfig{
		OnTime:       250 * time.Millisecond,
		OffTime:      750 * time.Millisecond,
		LoopTimes:    3,
		IntervalTime: 2 * time.Second,
	}

	payload, err := newLightCommand(ChamberLight, LightFlashing, cfg).Marshal()
	if err != nil {
		t.Fatalf("marshal light command: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal light command: %v", err)
	}
	want := map[string]any{
		"system": map[string]any{
			"command":       "ledctrl",
			"sequence_id":   "0",
			"led_node":      "chamber_light",
			"led_mode":      "flashing",
			"led_on_time":   float64(250),
			"led_off_time":  float64(750),
			"loop_times":    float64(3),
			"interval_time": float64(2000),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("light command = %#v, want %#v", got, want)
	}
}

func TestDefaultLightFlashingConfig(t *testing.T) {
	want := LightFlashingConfig{
		OnTime:       500 * time.Millisecond,
		OffTime:      500 * time.Millisecond,
		LoopTimes:    1,
		IntervalTime: time.Second,
	}
	if got := DefaultLightFlashingConfig(); got != want {
		t.Fatalf("default flashing config = %#v, want %#v", got, want)
	}
}
