package bambulabs_api

import (
	"context"

	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type CommandClient interface {
	Send(ctx context.Context, cmd *protocol.Command) error
}

type mqttCommandClient struct {
	mqtt *mqtt.MqttClient
}

func newMqttCommandClient(mqtt *mqtt.MqttClient) CommandClient {
	return &mqttCommandClient{
		mqtt: mqtt,
	}
}

func (c *mqttCommandClient) Send(
	ctx context.Context,
	cmd *protocol.Command,
) error {
	return c.mqtt.Publish(ctx, cmd)
}
