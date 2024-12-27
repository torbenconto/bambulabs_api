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
	if c.Command == "calibration" {
		return fmt.Sprintf(`{"%s": {"command": "%s", "option": "%s"}}`, c.Type, c.Command, c.Param)
	}
	if c.Type == System {
		return fmt.Sprintf(`{"%s": {"%s": "%s"}}`, c.Type, c.Command, c.Param)
	}
	return fmt.Sprintf(`{"%s": {"command": "%s", "param": "%s"}}`, c.Type, c.Command, c.Param)
}
