package bambulabs_api

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type lifetimePrinter struct {
	Printer
	closed     chan struct{}
	once       sync.Once
	closeCalls atomic.Int32
}

func (p *lifetimePrinter) Close() error {
	p.closeCalls.Add(1)
	p.once.Do(func() { close(p.closed) })
	return nil
}

func TestClientParentCancellationClosesManagedPrinters(t *testing.T) {
	parent, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	client := NewClient(parent)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	printers := map[string]*lifetimePrinter{
		"one": {closed: make(chan struct{})},
		"two": {closed: make(chan struct{})},
	}
	for serial, p := range printers {
		client.printers.Store(serial, p)
	}
	cancel()
	for serial, p := range printers {
		select {
		case <-p.closed:
		case <-time.After(time.Second):
			t.Fatalf("parent cancellation did not close printer %s", serial)
		}
	}
	require.Eventually(t, func() bool {
		_, one := client.printers.Load("one")
		_, two := client.printers.Load("two")
		return !one && !two
	}, time.Second, time.Millisecond)
	require.NoError(t, client.Close())
	for _, p := range printers {
		require.EqualValues(t, 1, p.closeCalls.Load())
	}
}

func TestClientExplicitCloseStillReleasesPrintersOnce(t *testing.T) {
	client := NewClient(t.Context())
	p := &lifetimePrinter{closed: make(chan struct{})}
	client.printers.Store("one", p)
	require.NoError(t, client.Close())
	require.NoError(t, client.Close())
	require.EqualValues(t, 1, p.closeCalls.Load())
	require.ErrorIs(t, client.ctx.Err(), context.Canceled)
	_, err := client.Load("one")
	require.ErrorIs(t, err, ErrPrinterNotFound)
}
