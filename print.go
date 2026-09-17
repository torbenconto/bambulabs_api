package bambulabs_api

import (
	"sync"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

// PrintState is the printer-reported state of a print operation.
type PrintState string

const (
	PrintStateUnknown PrintState = "UNKNOWN"
	PrintStateIdle    PrintState = "IDLE"
	PrintStatePrepare PrintState = "PREPARE"
	PrintStateRunning PrintState = "RUNNING"
	PrintStatePause   PrintState = "PAUSE"
	PrintStateFinish  PrintState = "FINISH"
	PrintStateFailed  PrintState = "FAILED"
)

// PrintInfo is a value snapshot of the latest print report.
type PrintInfo struct {
	State           PrintState
	TaskID          string
	SubtaskID       string
	Name            string
	Filename        string
	ProgressPercent int
	CurrentLayer    int
	TotalLayers     int
	Remaining       time.Duration
	FailureReason   string
}

// PrintSystem holds printer-reported print status.
type PrintSystem struct {
	mu   sync.RWMutex
	info PrintInfo
}

func NewPrintSystem() *PrintSystem {
	return &PrintSystem{info: PrintInfo{State: PrintStateUnknown}}
}

// Info returns an independent snapshot without requesting an update.
func (s *PrintSystem) Info() PrintInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.info
}

func (s *PrintSystem) apply(info PrintInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.info = info
}

type PrintDecoder struct{}

func NewPrintDecoder() *PrintDecoder {
	return &PrintDecoder{}
}

func (d *PrintDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil {
		return
	}
	r := report.Print
	p.Print().apply(PrintInfo{
		State:           PrintState(r.GcodeState),
		TaskID:          r.TaskID,
		SubtaskID:       r.SubtaskID,
		Name:            r.SubtaskName,
		Filename:        r.GcodeFile,
		ProgressPercent: r.MCPercent,
		CurrentLayer:    r.LayerNum,
		TotalLayers:     r.TotalLayerNum,
		Remaining:       time.Duration(r.MCRemainingTime) * time.Minute,
		FailureReason:   r.FailReason,
	})
}
