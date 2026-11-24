package bambulabs_api_test

import (
	"context"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/emulator"
)

func TestFilesAfterOptionalFTPFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	// Accept and close locally so the optional TLS connection reliably fails,
	// without depending on an unused fixed port or any external FTP server.
	ftpListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ftpListener.Close() })
	go func() {
		if conn, err := ftpListener.Accept(); err == nil {
			_ = conn.Close()
		}
	}()
	cfg := &bambulabs_api.Config{
		Host: net.ParseIP("127.0.0.1"), FTPPort: ftpListener.Addr().(*net.TCPAddr).Port,
		Model: bambulabs_api.ModelA1, AccessCode: "test", SerialNumber: "FTPFAIL0001",
	}
	emu, err := emulator.Start(ctx, cfg, 0, filepath.Join("fixtures", "mock", "all_fans.json"))
	require.NoError(t, err)
	t.Cleanup(emu.Stop)
	cfg.MQTTPort = emu.Port()
	p, err := bambulabs_api.NewPrinter(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, p.Close()) })

	files := p.Files()
	require.NotNil(t, files, "degraded printers must still expose the FileClient contract")
	entries, err := files.List("/")
	require.Nil(t, entries)
	require.ErrorIs(t, err, bambulabs_api.ErrFTPUnavailable)
	require.ErrorIs(t, files.Download("/file", io.Discard), bambulabs_api.ErrFTPUnavailable)
	require.ErrorIs(t, files.Upload("/file", strings.NewReader("data")), bambulabs_api.ErrFTPUnavailable)
	require.ErrorIs(t, files.Delete("/file"), bambulabs_api.ErrFTPUnavailable)
	// MQTT remains usable after the non-vital connection fails.
	require.NoError(t, p.RequestUpdate(ctx))
}
