package bambulabs_api_test

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/emulator"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func newEmulatedConfig(t *testing.T, model bambulabs_api.Model) *bambulabs_api.Config {
	t.Helper()
	return &bambulabs_api.Config{
		Host:         net.ParseIP("127.0.0.1"),
		MQTTPort:     freePort(t),
		Model:        model,
		AccessCode:   "test",
		SerialNumber: "EMULATOR0001",
	}
}

func startEmulatedPrinter[P any](
	t *testing.T,
	ctor func(context.Context, *bambulabs_api.Config) (P, error),
	model bambulabs_api.Model,
	reportFile string,
) (P, *emulator.Emulator) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	cfg := newEmulatedConfig(t, model)

	emu, err := emulator.Start(ctx, cfg, cfg.MQTTPort, filepath.Join("fixtures", reportFile))
	require.NoError(t, err)
	t.Cleanup(emu.Stop)

	p, err := ctor(ctx, cfg)
	require.NoError(t, err)

	if closer, ok := any(p).(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}

	return p, emu
}
