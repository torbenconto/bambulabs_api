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

func startEmulatedPrinter(t *testing.T, model bambulabs_api.Model, reportFile string) (bambulabs_api.Printer, *emulator.Emulator) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	cfg := &bambulabs_api.Config{
		Host:         net.ParseIP("127.0.0.1"),
		Model:        model,
		AccessCode:   "test",
		SerialNumber: "EMULATOR0001",
	}
	emu, err := emulator.Start(ctx, cfg, 0, filepath.Join("fixtures", reportFile))
	require.NoError(t, err)
	t.Cleanup(emu.Stop)
	cfg.MQTTPort = emu.Port()

	p, err := bambulabs_api.NewPrinter(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, p.Close()) })
	return p, emu
}
