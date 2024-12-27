package mqtt

const (
	LightOn  = `"{system": {"light_mode": "on"}}`
	LightOff = `{"system": {"light_mode": "off"}}`
)

const (
	GetVersion = `{"info": {"get_version"}}`
)

const (
	Pause  = `"{print": {"command": "pause"}}`
	Resume = `"{print": {"command": "resume"}}`
	Stop   = `"{print": {"command": "stop"}}`
)

const (
	PushAll = `"{pushing": {"command": "push_all"}}`
)
