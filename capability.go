package bambulabs_api

type Capability uint8

// We derive light and fan capabilites seperately
const (
	CapabilityCamera Capability = 1 << iota
	CapabilityAms
)

var allCapabilities Capability = CapabilityCamera | CapabilityAms

var capabilitiesForModel = map[Model]Capability{
	ModelUnknown: 0,
	ModelA1Mini:  allCapabilities,
	ModelA1:      allCapabilities,
	ModelP1S:     CapabilityAms,
	ModelP2S:     allCapabilities,
	ModelX1C:     allCapabilities,
	ModelX1E:     allCapabilities,
	ModelX2D:     allCapabilities,
	ModelH2:      allCapabilities,
	ModelH2S:     allCapabilities,
	ModelH2D:     allCapabilities,
	ModelH2DPro:  allCapabilities,
	ModelH2C:     allCapabilities,
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
