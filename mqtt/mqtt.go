package mqtt

import (
	"crypto/tls"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
	"time"
)

const (
	Clientid = "golang-bambulabs-api"
)

type ClientConfig struct {
	Host       string
	Port       int
	Serial     string
	AccessCode string
	Username   string
	Timeout    time.Duration
}

type Client struct {
	config ClientConfig
	client paho.Client
}

func NewClient(config ClientConfig) *Client {
	options := paho.NewClientOptions()
	// Maybe tls:// or tcp://
	options.AddBroker("tls://" + config.Host + ":" + strconv.Itoa(config.Port))
	options.SetClientID(Clientid)
	options.SetUsername(config.Username)
	options.SetPassword(config.AccessCode)
	options.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	options.SetAutoReconnect(true)

	options.SetOnConnectHandler(func(client paho.Client) {
		topic := fmt.Sprintf("/device/%s/report", config.Serial)
		if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			log.Printf("Error subscribing to topic %s: %s", topic, token.Error())
		} else {
			log.Printf("Subscribed to topic %s", topic)
		}
	})
	options.SetConnectionLostHandler(func(client paho.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	return &Client{
		config: config,
		client: paho.NewClient(&paho.ClientOptions{}),
	}
}

func (c *Client) Connect() error {
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) Disconnect() {
	// Allow 250ms for in-flight messages
	c.client.Disconnect(250)

	log.Println("MQTT client disconnected")
}
