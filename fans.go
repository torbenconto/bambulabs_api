package bambulabs_api

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Fan int

const (
	PartCoolingFan        Fan = 1
	AuxillaryFan          Fan = 2
	ChamberFan            Fan = 3
	SecondaryAuxillaryFan Fan = 10
)

type FanInfo struct {
	Fan     Fan
	Percent int
}

type FanSystem struct {
	sendGate      chan struct{}
	revisions     map[Fan]uint64
	mu            sync.RWMutex
	fans          map[Fan]FanInfo
	commandClient CommandClient
}

func NewFanSystem(commandClient CommandClient) *FanSystem {
	return &FanSystem{
		sendGate:      make(chan struct{}, 1),
		revisions:     make(map[Fan]uint64),
		fans:          make(map[Fan]FanInfo),
		commandClient: commandClient,
	}
}

// Get returns the latest reported or optimistically requested state.
// It returns [ErrFanUnavailable] if the printer hasn't reported this fan yet (e.g. an auxiliary fan not physically installed).
func (f *FanSystem) Get(id Fan) (FanInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	info, ok := f.fans[id]
	if !ok {
		return FanInfo{}, ErrFanUnavailable
	}

	return info, nil
}

// Set updates local state immediately, then sends the command. A failed send
// restores the previous state unless a printer report has superseded it.
// Later reports remain authoritative. Percent must be in [0, 100] and is
// rounded to the nearest 10, matching the command sent to the printer.
func (f *FanSystem) Set(ctx context.Context, id Fan, percent int) error {
	ctx, cancel := withDefaultOpTimeout(ctx)
	defer cancel()

	if percent < 0 || percent > 100 {
		return ErrInvalidFanPercent
	}

	select {
	case f.sendGate <- struct{}{}:
		defer func() { <-f.sendGate }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	percent = int(math.Round(float64(percent)/10)) * 10

	f.mu.Lock()
	previous, ok := f.fans[id]
	if !ok {
		f.mu.Unlock()
		return ErrFanUnavailable
	}
	f.revisions[id]++
	revision := f.revisions[id]
	f.fans[id] = FanInfo{Fan: id, Percent: percent}
	f.mu.Unlock()

	if err := f.commandClient.Send(ctx, newFanCommand(id, percent)); err != nil {
		f.mu.Lock()
		if f.revisions[id] == revision {
			f.fans[id] = previous
		}
		f.mu.Unlock()
		return err
	}
	return nil
}

// apply records a fan state reported by the printer. Called by [FanDecoder]
// while holding the printer's decode lock.
func (f *FanSystem) apply(id Fan, percent int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.revisions[id]++
	f.fans[id] = FanInfo{
		Fan:     id,
		Percent: percent,
	}
}

func newFanCommand(id Fan, percent int) *protocol.Command {
	speed := percentToGCodeSpeed(percent)

	return protocol.NewCommand(protocol.Print).
		WithCommand("gcode_line").
		WithParam(fmt.Sprintf("M106 P%d S%d\n", id, speed))
}

type FanDecoder struct{}

func NewFanDecoder() *FanDecoder {
	return &FanDecoder{}
}

func (f *FanDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil {
		return
	}

	if report.Print.AuxPartFan {
		p.cap.Add(CapabilityAuxFan)
	}
	// Missing fields in a delta report must not reset existing state.
	if report.Print.BigFan1Speed != "" && p.cap.Has(CapabilityAuxFan) {
		p.Fans().apply(AuxillaryFan, parsePercent(report.Print.BigFan1Speed))
	}
	if report.Print.CoolingFanSpeed != "" {
		p.Fans().apply(PartCoolingFan, parsePercent(report.Print.CoolingFanSpeed))
	}
	if report.Print.BigFan2Speed != "" {
		p.Fans().apply(ChamberFan, parsePercent(report.Print.BigFan2Speed))
	}
}
