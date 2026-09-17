package bambulabs_api

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

func TestPrintDecoderFixtures(t *testing.T) {
	for _, tc := range []struct {
		file  string
		model Model
		want  PrintInfo
	}{
		{"a1.json", ModelA1, PrintInfo{
			State: PrintStateIdle, TaskID: "0", SubtaskID: "0",
		}},
		{"h2dpro.json", ModelH2DPro, PrintInfo{
			State: PrintStateIdle, FailureReason: "0",
		}},
		{"p1p_no_ams.json", ModelP1P, PrintInfo{
			State: PrintStateFinish, TaskID: "715417865", SubtaskID: "715417867",
			Name: "sensor_altanlås_v2_v3", ProgressPercent: 100,
			CurrentLayer: 124, TotalLayers: 124,
		}},
		{"x2d.json", ModelX2D, PrintInfo{
			State: PrintStateRunning, TaskID: "904240393", SubtaskID: "904240393",
			Name: "Build Plate Cleaner Trays - A1 Mini", Filename: "/data/Metadata/plate_3.gcode",
			ProgressPercent: 98, CurrentLayer: 129, TotalLayers: 140,
			Remaining: 2 * time.Minute, FailureReason: "0",
		}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			p := newTestPrinter(t, tc.model, tc.file)
			require.Equal(t, tc.want, p.Print().Info())
		})
	}
}

func TestPrintDecoderNoPrintReport(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "")
	require.Equal(t, PrintInfo{State: PrintStateUnknown}, p.Print().Info())
	NewPrintDecoder().Apply(p, &protocol.Report{})
	require.Equal(t, PrintInfo{State: PrintStateUnknown}, p.Print().Info())
}

func TestPrintSystemReplacesSnapshot(t *testing.T) {
	p := newTestPrinter(t, ModelX2D, "x2d.json")
	running := p.Print().Info()
	idle := newTestPrinter(t, ModelA1, "a1.json").Print().Info()
	p.Print().apply(idle)
	require.Equal(t, idle, p.Print().Info(), "zero values replace previous job data")
	require.Equal(t, 98, running.ProgressPercent, "previous snapshot is independent")

	copy := p.Print().Info()
	copy.TaskID = "caller mutation"
	require.Equal(t, idle, p.Print().Info())
}

func TestPrintSystemConcurrentSnapshots(t *testing.T) {
	running := newTestPrinter(t, ModelX2D, "x2d.json").Print().Info()
	idle := newTestPrinter(t, ModelA1, "a1.json").Print().Info()
	system := NewPrintSystem()
	system.apply(running)
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			system.apply(idle)
			system.apply(running)
		}
	})
	for range 3 {
		wg.Go(func() {
			for range 100 {
				info := system.Info()
				assert.True(t, info == idle || info == running, "snapshot must be coherent")
				info.Filename = "caller copy"
			}
		})
	}
	wg.Wait()
	require.Equal(t, running, system.Info())
}
