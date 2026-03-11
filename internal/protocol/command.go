package protocol

import "encoding/json"

type MessageType string

const (
	Print   MessageType = "print"
	System  MessageType = "system"
	Pushing MessageType = "pushing"
	Upgrade MessageType = "upgrade"
	Info    MessageType = "info"
)

// Command is a wrapper for an MQTT command to the printer to ensure proper structure
type Command struct {
	messageType MessageType
	id          string
	fields      map[string]any
}

func NewCommand(messageType MessageType) *Command {
	c := &Command{
		messageType: messageType,
		id:          "0",
		fields:      make(map[string]any),
	}

	return c.WithSequenceID(c.id)
}

func (c *Command) Set(key string, value any) *Command {
	c.fields[key] = value
	return c
}

func (c *Command) WithCommand(cmd any) *Command {
	return c.Set("command", cmd)
}

func (c *Command) WithParam(param any) *Command {
	return c.Set("param", param)
}

func (c *Command) WithSequenceID(id string) *Command {
	return c.Set("sequence_id", id)
}

func (c *Command) Marshal() ([]byte, error) {
	return json.Marshal(map[string]any{
		string(c.messageType): c.fields,
	})
}
