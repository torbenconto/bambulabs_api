package bambulabs_api

type Light string

const (
	ChamberLight Light = "chamber_light"
	WorkLight    Light = "work_light" // Defined but seemingly unused
)

type LightMode string

const (
	LightOn       LightMode = "on"
	LightOff      LightMode = "off"
	LightFlashing LightMode = "flashing"
)

// I know what you're thinking but hear me out, the open-air printers use ChamberLight for their WorkLight which makes no sense
// WorkLight is defined but not used so it'll likely change in the future
