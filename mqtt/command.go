package mqtt

import (
	"encoding/json"
)

type MessageType string

const (
	Print   MessageType = "print"
	System              = "system"
	Pushing             = "pushing"
)

type Command struct {
	Type MessageType
	id   string

	fields map[string]interface{}
}

func NewCommand(msgType MessageType) *Command {
	cmd := &Command{
		Type:   msgType,
		id:     "0",
		fields: make(map[string]interface{}),
	}

	return cmd.AddIdField(cmd.id)

}

// AddField adds a field with the given key and value.
func (c *Command) AddField(key string, value interface{}) *Command {
	c.fields[key] = value

	return c
}

// AddCommandField adds a field with key "command" and the given value.
func (c *Command) AddCommandField(value interface{}) *Command {
	c.AddField("command", value)

	return c
}

// AddParamField adds a field with key "param" and the given value.
func (c *Command) AddParamField(value interface{}) *Command {
	c.AddField("param", value)

	return c
}

func (c *Command) AddIdField(id string) *Command {
	c.AddField("sequence_id", id)

	return c
}

// JSON returns the command as a JSON string.
func (c *Command) JSON() (string, error) {
	data := make(map[string]interface{})
	for k, v := range c.fields {
		data[k] = v
	}
	message := map[string]interface{}{
		string(c.Type): data,
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
