package bambulabs_api

import "github.com/torbenconto/bambulabs_api/hms"

type State struct {
	cap Capability // private for now to prevent user re-assignment

	Errors []hms.Error

	AMS       AMSSystem
	Extruders ExtruderSystem
	Nozzles   NozzleSystem
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
