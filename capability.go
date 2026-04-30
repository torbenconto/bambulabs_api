package bambulabs_api

type Capability uint64

// We derive light and fan capabilites seperately
const (
	CapabilityCamera Capability = 1 << iota
	CapabilityAms
)

var capabilitiesForModel = map[Model]Capability{
	ModelUnknown: 0,
	ModelA1Mini:  CapabilityCamera | CapabilityAms,
	ModelA1:      CapabilityCamera | CapabilityAms,
	ModelP1S:     CapabilityAms,
	ModelP2S:     CapabilityCamera | CapabilityAms,
	ModelX1C:     CapabilityCamera | CapabilityAms,
	ModelX1E:     CapabilityCamera | CapabilityAms,
	ModelX2D:     CapabilityCamera | CapabilityAms,
	ModelH2:      CapabilityCamera | CapabilityAms,
	ModelH2S:     CapabilityCamera | CapabilityAms,
	ModelH2D:     CapabilityCamera | CapabilityAms,
	ModelH2DPro:  CapabilityCamera | CapabilityAms,
	ModelH2C:     CapabilityCamera | CapabilityAms,
}

func (c Capability) Has(cap Capability) bool {
	return c&cap != 0
}

func CapabilityForModel(m Model) Capability {
	return capabilitiesForModel[m]
}

func HasCapability(m Model, cap Capability) bool {
	return CapabilityForModel(m)&cap != 0
}
