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
