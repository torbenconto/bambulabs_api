package bambulabs_api

type NozzleType uint8

const (
	NozzleTypeUndefined       NozzleType = 0
	NozzleTypeHardenedSteel   NozzleType = 1
	NozzleTypeStainlessSteel  NozzleType = 2
	NozzleTypeTungstenCarbide NozzleType = 3
	NozzleTypeBrass           NozzleType = 4
	NozzleTypeE3D             NozzleType = 5
)

type NozzleFlowType uint8

const (
	NozzleFlowTypeNone        NozzleFlowType = iota
	NozzleFlowTypeStandard                   // S_FLOW - standard 1.75mm
	NozzleFlowTypeHighFlow                   // H_FLOW - high flow
	NozzleFlowTypeTPUHighFlow                // U_FLOW - TPU high flow
)

type NozzleDiameterType uint8

const (
	NozzleDiameterTypeNone NozzleDiameterType = iota
	NozzleDiameter0_2
	NozzleDiameter0_4
	NozzleDiameter0_6
	NozzleDiameter0_8
)

type Nozzle struct {
	ID            int            // Position ID (0-0x0F for extruder, 0x10-0x1F for rack)
	Type          NozzleType     // Material type
	FlowType      NozzleFlowType // Standard or High-Flow
	Diameter      float32        // possibly NozzleDiameterType??
	Wear          float32        // Wear percentage (0.0-1.0)
	PrintTime     int            // Total print time with this nozzle
	SerialNumber  string         // Firmware SN
	FilamentID    string         // Last filament used
	FilamentColor string         // Last filament color
	Status        int            // Status bits
	OnRack        bool           // Is this nozzle on the rack?
}

type NozzleSystem struct {
	// Nozzles installed on extruders (0x00-0x0F)
	extruderNozzles map[int]*Nozzle

	// Nozzles on the rack (0x10-0x1F, need -0x10 to access map)
	rackNozzles map[int]*Nozzle

	// State
	isIdle       bool
	isRefreshing bool

	replaceNozzleSrc *int // Source nozzle being replaced
	replaceNozzleTar *int // Target position for replacement
}

func NewNozzleSystem() *NozzleSystem {
	return &NozzleSystem{
		extruderNozzles: make(map[int]*Nozzle),
		rackNozzles:     make(map[int]*Nozzle),
		isIdle:          true,
	}
}

func (ns *NozzleSystem) GetExtruderNozzle(id int) *Nozzle {
	if nozzle, ok := ns.extruderNozzles[id]; ok {
		return nozzle
	}
	return nil
}
