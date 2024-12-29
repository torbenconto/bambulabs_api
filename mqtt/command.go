package mqtt

import "fmt"

type MessageType string

const (
	Print   MessageType = "print"
	System              = "system"
	Pushing             = "pushing"
)

type Command struct {
	Type    MessageType
	Command string
	Param   string
}

func (c Command) JSON() string {
	format := `{"%s": {"command": "%s", "param": "%s"}}`
	if c.Command == "calibration" {
		format = `{"%s": {"command": "%s", "option": "%s"}}`
	} else if c.Type == System {
		format = `{"%s": {"%s": "%s"}}`
	}
	return fmt.Sprintf(format, c.Type, c.Command, c.Param)
}
