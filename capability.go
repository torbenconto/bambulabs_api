package bambulabs_api

// Capability represents the **current** ability of a printer to perform some action.
// These are dynamically inferred based on telemetry data and stored within the state of a [Printer]
type Capability uint64

const (
	CapabilityAMS Capability = 1 << iota
	CapabilityLidar
	CapabilityDualExtruder
	CapabilityFilamentSwitcher
	CapabilityLazer
)

func (c *Capability) Add(cap Capability) {
	*c |= cap
}

func (c *Capability) Remove(cap Capability) {
	*c &^= cap
}

func (c Capability) Has(cap Capability) bool {
	return c&cap != 0
}

func (c Capability) HasAll(caps ...Capability) bool {
	for _, cap := range caps {
		if !c.Has(cap) {
			return false
		}
	}
	return true
}

func (c Capability) HasAny(caps ...Capability) bool {
	for _, cap := range caps {
		if c.Has(cap) {
			return true
		}
	}
	return false
}
