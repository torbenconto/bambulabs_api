package bambulabs_api

type ExtruderID uint8

const (
	MainExtruder    ExtruderID = 0
	DeputyExtruder  ExtruderID = 1
	InvalidExtruder ExtruderID = 0xF
)

type ExtruderSystem struct {
	numExtruders      int
	currID            ExtruderID
	targetID          ExtruderID
	loadingExtruderID ExtruderID
	busyLoading       bool

	extruders []*Extruder
}

type FilamentSlot struct {
	AMSId  uint8
	SlotId uint8
}

type Extruder struct {
	ID ExtruderID

	// Physical state
	HasFilament       bool
	BufferHasFilament bool
	HasNozzle         bool

	// Temperature
	CurrentTemp float32
	TargetTemp  float32

	// Current nozzle
	NozzleID uint8

	// Filament routing
	CurrentSlot  FilamentSlot // snow
	TargetSlot   FilamentSlot // star
	PreviousSlot FilamentSlot // spre

	// Filament change
	// FilamentChangeStep *FilamentStep

	// AMS state
	// AMSStatus  uint16
	// RFIDStatus uint16

	// Backup filament trays
	BackupSlots []uint32
}
