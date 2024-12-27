package bambulabs_api

import (
	"bambulabs-api/mqtt"
	"net"
)

type Printer struct {
	IpAddr     net.IP
	AccessCode string
	Serial     string

	MQTTClient *mqtt.Client
}

func NewPrinter(IpAddr net.IP, AccessCode, Serial string) *Printer {
	return &Printer{
		IpAddr:     IpAddr,
		AccessCode: AccessCode,
		Serial:     Serial,

		MQTTClient: mqtt.NewClient(&mqtt.ClientConfig{
			Host:       IpAddr,
			Port:       8883,
			Serial:     Serial,
			Username:   "bblp",
			AccessCode: AccessCode,
		}),
	}
}

func (p *Printer) Connect() error {
	return p.MQTTClient.Connect()
}

func (p *Printer) Disconnect() {
	p.MQTTClient.Disconnect()
}

func (p *Printer) LightOn() error {
	return p.MQTTClient.Publish(mqtt.LightOn)
}

func (p *Printer) LightOff() error {
	return p.MQTTClient.Publish(mqtt.LightOff)
}
