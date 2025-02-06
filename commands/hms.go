package commands

import (
	"github.com/torbenconto/bambulabs_api/hms"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
)

type HMS struct {
	mqttClient *mqtt.Client
}

func CreateHMSInstance(mqttClient *mqtt.Client) *HMS {
	return &HMS{mqttClient: mqttClient}
}

func (p *HMS) GetErrors() []hms.Error {
	return p.mqttClient.Data().Print.HMS
}
