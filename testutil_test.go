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

type fakeCommandClient struct{}

func (fakeCommandClient) Send(context.Context, *protocol.Command) error {
	return nil
}

type capturingCommandClient struct {
	mu       sync.Mutex
	commands []*protocol.Command
}

func (c *capturingCommandClient) Send(_ context.Context, cmd *protocol.Command) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.commands = append(c.commands, cmd)
	return nil
}

func (c *capturingCommandClient) last(t *testing.T) map[string]any {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	require.NotEmpty(t, c.commands, "no command was sent")

	payload, err := c.commands[len(c.commands)-1].Marshal()
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(payload, &got))
	return got
}

func newTestPrinter(tb testing.TB, model Model, reportFile string) *printer {
	tb.Helper()

	p := &printer{
		cfg: Config{
			Model: model,
		},
		AMS:    NewAMSSystem(),
		Lights: NewLightSystem(fakeCommandClient{}),
		Fans:   NewFanSystem(fakeCommandClient{}),
	}

	p.decoder = *NewDecoder(model)

	if reportFile != "" {
		data, err := os.ReadFile(filepath.Join("fixtures", reportFile))
		if err != nil {
			tb.Fatal(err)
		}

		var report protocol.Report
		if err := json.Unmarshal(data, &report); err != nil {
			tb.Fatal(err)
		}

		p.decoder.Apply(p, &report)
	}

	return p
}
