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

	fields map[string]interface{}
}

func NewCommand(msgType MessageType) *Command {
	return &Command{
		Type:   msgType,
		fields: make(map[string]interface{}),
	}
}

func (c *Command) AddField(key string, value interface{}) *Command {
	c.fields[key] = value

	return c
}

func (c *Command) AddCommandField(value string) *Command {
	c.AddField("command", value)

	return c
}

func (c *Command) AddParamField(value string) *Command {
	c.AddField("param", value)

	return c
}

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
