package bambulabs_api

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
