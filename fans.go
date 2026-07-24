package bambulabs_api

import (
	"context"
	"fmt"
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
	mu            sync.RWMutex
	fans          map[Fan]FanInfo
	commandClient CommandClient
}

func NewFanSystem(commandClient CommandClient) *FanSystem {
	return &FanSystem{
		fans:          make(map[Fan]FanInfo),
		commandClient: commandClient,
	}
}

// Get returns the last known state of the given fan, as reported by the
// printer. It returns ErrFanUnavalible if the printer hasn't reported this
// fan yet (e.g. an auxiliary fan not physically installed).
func (f *FanSystem) Get(id Fan) (FanInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	info, ok := f.fans[id]
	if !ok {
		return FanInfo{}, ErrFanUnavailable
	}

	return info, nil
}

func (f *FanSystem) Set(ctx context.Context, id Fan, percent int) error {
	if _, err := f.Get(id); err != nil {
		return err
	}

	return f.commandClient.Send(ctx, newFanCommand(id, percent))
}

// apply records a fan state reported by the printer. Called by [FanDecoder]
// while holding the printer's decode lock.
func (f *FanSystem) apply(id Fan, percent int) {
	f.mu.Lock()
	defer f.mu.Unlock()

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
		p.Fans.apply(AuxillaryFan, parsePercent(report.Print.BigFan1Speed))
	}

	p.Fans.apply(PartCoolingFan, parsePercent(report.Print.CoolingFanSpeed))
	p.Fans.apply(ChamberFan, parsePercent(report.Print.BigFan2Speed))
}
