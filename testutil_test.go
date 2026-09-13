package bambulabs_api

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

// commandClientFunc lets individual tests control send failures and timing.
type commandClientFunc func(context.Context, *protocol.Command) error

func (fn commandClientFunc) Send(ctx context.Context, cmd *protocol.Command) error {
	return fn(ctx, cmd)
}

type capturingCommandClient struct {
	mu       sync.Mutex
	payloads [][]byte
}

func (c *capturingCommandClient) Send(_ context.Context, cmd *protocol.Command) error {
	payload, err := cmd.Marshal()
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.payloads = append(c.payloads, payload)
	return nil
}

func (c *capturingCommandClient) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.payloads)
}

func (c *capturingCommandClient) last(tb testing.TB) map[string]any {
	tb.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	require.NotEmpty(tb, c.payloads, "no command was sent")

	var got map[string]any
	require.NoError(tb, json.Unmarshal(c.payloads[len(c.payloads)-1], &got))

	return got
}

func newTestPrinter(tb testing.TB, model Model, reportFile string) *printer {
	tb.Helper()

	commands := &capturingCommandClient{}
	p := &printer{
		cfg:         Config{Model: model},
		amsSystem:   NewAMSSystem(),
		lightSystem: NewLightSystem(commands),
		fanSystem:   NewFanSystem(commands),
		hmsSystem:   NewHMSSystem(),
		decoder:     *NewDecoder(model),
		ready:       make(chan struct{}),
	}
	if reportFile != "" {
		data, err := os.ReadFile(filepath.Join("fixtures", reportFile))
		require.NoError(tb, err)
		applyTestReport(tb, p, string(data))
	}
	return p
}

func applyTestReport(tb testing.TB, p *printer, payload string) {
	tb.Helper()
	require.True(tb, json.Valid([]byte(payload)), "invalid test report: %s", payload)
	p.updateState([]byte(payload))
}
