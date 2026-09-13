package bambulabs_api

import (
	"context"
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

func TestUpdateStateReadiness(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "")
	for _, payload := range []string{"invalid", "{}", `{"system":{"command":"ledctrl"}}`} {
		p.updateState([]byte(payload))
		select {
		case <-p.ready:
			t.Fatalf("printer became ready on non-telemetry payload %q", payload)
		default:
		}
	}
	applyTestReport(t, p, `{"print":{"cooling_fan_speed":"0"}}`)
	select {
	case <-p.ready:
	default:
		t.Fatal("printer did not become ready after telemetry")
	}
}
