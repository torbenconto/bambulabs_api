package bambulabs_api

import "github.com/torbenconto/bambulabs_api/hms"

type State struct {
	cap Capability

	Errors []hms.Error

	AMS       AMSSystem
	Extruders ExtruderSystem
	Nozzles   NozzleSystem
}

func NewState() *State {
	return &State{
		Nozzles: *NewNozzleSystem(),
	}
}

func (s State) Capability() Capability {
	return s.cap
}

// Get nozzle installed on an extruder
func (s *State) GetExtruderNozzle(extruderID ExtruderID) *Nozzle {
	extruder := s.Extruders.extruders[extruderID]
	if extruder == nil {
		return nil
	}

	// NozzleID is an index (0-0x0F for extruder nozzles)
	return s.Nozzles.GetExtruderNozzle(int(extruder.NozzleID))
}
