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

// PrintInfo is a value snapshot of the latest reported print fields.
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
	p.Print().applyReport(report.Print)
}

func (s *PrintSystem) applyReport(r *protocol.PrintReport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info := s.info
	// Task identifiers can arrive in separate deltas too. Only reported
	// fields replace state; an identifier change does not clear other fields.
	if r.GcodeState != nil {
		info.State = PrintState(*r.GcodeState)
	}
	if r.TaskID != nil {
		info.TaskID = *r.TaskID
	}
	if r.SubtaskID != nil {
		info.SubtaskID = *r.SubtaskID
	}
	if r.SubtaskName != nil {
		info.Name = *r.SubtaskName
	}
	if r.GcodeFile != nil {
		info.Filename = *r.GcodeFile
	}
	if r.MCPercent != nil {
		info.ProgressPercent = *r.MCPercent
	}
	if r.LayerNum != nil {
		info.CurrentLayer = *r.LayerNum
	}
	if r.TotalLayerNum != nil {
		info.TotalLayers = *r.TotalLayerNum
	}
	if r.MCRemainingTime != nil {
		info.Remaining = time.Duration(*r.MCRemainingTime) * time.Minute
	}
	if r.FailReason != nil {
		info.FailureReason = *r.FailReason
	}
	s.info = info
}
