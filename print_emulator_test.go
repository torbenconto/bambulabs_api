package bambulabs_api_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bambu "github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestPrintSystemEmulator(t *testing.T) {
	p, emu := startEmulatedPrinter(t, bambu.ModelX2D, "x2d.json")
	running := bambu.PrintInfo{
		State: bambu.PrintStateRunning, TaskID: "904240393", SubtaskID: "904240393",
		Name: "Build Plate Cleaner Trays - A1 Mini", Filename: "/data/Metadata/plate_3.gcode",
		ProgressPercent: 98, CurrentLayer: 129, TotalLayers: 140,
		Remaining: 2 * time.Minute, FailureReason: "0",
	}
	require.Equal(t, running, p.Print().Info())

	for _, tc := range []struct {
		file string
		want bambu.PrintInfo
	}{
		{"p1p_no_ams.json", bambu.PrintInfo{
			State: bambu.PrintStateFinish, TaskID: "715417865", SubtaskID: "715417867",
			Name: "sensor_altanlås_v2_v3", ProgressPercent: 100,
			CurrentLayer: 124, TotalLayers: 124,
		}},
		{"a1.json", bambu.PrintInfo{
			State: bambu.PrintStateIdle, TaskID: "0", SubtaskID: "0",
		}},
		{"x2d.json", running},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("fixtures", tc.file))
			require.NoError(t, err)
			var report protocol.Report
			require.NoError(t, json.Unmarshal(data, &report))
			require.NoError(t, emu.PublishReport(report))
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				assert.Equal(c, tc.want, p.Print().Info())
			}, 2*time.Second, 10*time.Millisecond)
		})
	}
}
