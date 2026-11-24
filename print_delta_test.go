package bambulabs_api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintDecoderPreservesFieldsAbsentFromDelta(t *testing.T) {
	p := newTestPrinter(t, ModelX2D, "x2d.json")
	want := p.Print().Info()
	applyTestReport(t, p, `{"print":{"cooling_fan_speed":"6","hms":[]}}`)
	require.Equal(t, want, p.Print().Info(), "unrelated telemetry must not erase the current job")
	applyTestReport(t, p, `{"print":{"mc_percent":99}}`)
	want.ProgressPercent = 99
	require.Equal(t, want, p.Print().Info(), "a progress delta must preserve names, layers, and state")
	applyTestReport(t, p, `{"print":{"gcode_state":"PAUSE","subtask_name":"renamed"}}`)
	want.State = PrintStatePause
	want.Name = "renamed"
	require.Equal(t, want, p.Print().Info())
}

func TestPrintDecoderAppliesExplicitZeroAndEmptyValues(t *testing.T) {
	p := newTestPrinter(t, ModelX2D, "x2d.json")
	applyTestReport(t, p, `{"print":{
		"gcode_state":"IDLE", "task_id":"", "subtask_id":"", "subtask_name":"", "gcode_file":"",
		"mc_percent":0, "layer_num":0, "total_layer_num":0, "mc_remaining_time":0, "fail_reason":""
	}}`)
	require.Equal(t, PrintInfo{State: PrintStateIdle}, p.Print().Info())
}

func TestPrintDecoderKeepsUnknownStateUntilReported(t *testing.T) {
	p := newTestPrinter(t, ModelA1, "")
	applyTestReport(t, p, `{"print":{"cooling_fan_speed":"0"}}`)
	require.Equal(t, PrintInfo{State: PrintStateUnknown}, p.Print().Info())
}

func TestPrintDecoderRetainsFieldsAbsentFromTaskChange(t *testing.T) {
	p := newTestPrinter(t, ModelX2D, "x2d.json")
	want := p.Print().Info()
	applyTestReport(t, p, `{"print":{"task_id":"next-task","subtask_id":"next-plate","gcode_state":"PREPARE"}}`)
	want.State = PrintStatePrepare
	want.TaskID = "next-task"
	want.SubtaskID = "next-plate"
	require.Equal(t, want, p.Print().Info())
}

func TestPrintDecoderSplitTaskIdentity(t *testing.T) {
	for _, taskFirst := range []bool{true, false} {
		name := "subtask then task"
		if taskFirst {
			name = "task then subtask"
		}
		t.Run(name, func(t *testing.T) {
			p := newTestPrinter(t, ModelX2D, "x2d.json")
			if taskFirst {
				applyTestReport(t, p, `{"print":{"task_id":"next-task","gcode_state":"PREPARE"}}`)
				applyTestReport(t, p, `{"print":{"subtask_id":"next-plate"}}`)
			} else {
				applyTestReport(t, p, `{"print":{"subtask_id":"next-plate","gcode_state":"PREPARE"}}`)
				applyTestReport(t, p, `{"print":{"task_id":"next-task"}}`)
			}
			info := p.Print().Info()
			require.Equal(t, "next-task", info.TaskID)
			require.Equal(t, "next-plate", info.SubtaskID)
			require.Equal(t, PrintStatePrepare, info.State)
		})
	}
}
