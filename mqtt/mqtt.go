package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
)

// https://github.com/Doridian/OpenBambuAPI/blob/main/mqtt.md

import (
	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	clientid     = "golang-bambulabs-api"
	topic        = "device/%s/report"
	commandTopic = "device/%s/request"
)

type ClientConfig struct {
	Host       net.IP
	Port       int
	Serial     string
	Username   string
	AccessCode string

	// Duration before data is re-fetched
	Timeout time.Duration
}

type Client struct {
	config *ClientConfig
	client paho.Client

	mutex      sync.Mutex
	data       Message
	lastUpdate time.Time
}

func NewClient(config *ClientConfig) *Client {
	options := paho.NewClientOptions()
	// Maybe tls:// or tcp://
	options.AddBroker("mqtts://" + config.Host.String() + ":" + strconv.Itoa(config.Port))
	options.SetClientID(clientid)
	options.SetUsername(config.Username)
	options.SetPassword(config.AccessCode)
	options.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	options.SetAutoReconnect(true)

	c := &Client{
		config:     config,
		data:       Message{},
		lastUpdate: time.Now(),
	}

	// TODO: needs improvement
	options.SetOnConnectHandler(func(client paho.Client) {
		topic := fmt.Sprintf(topic, config.Serial)
		if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			log.Printf("Error subscribing to topic %s: %s", topic, token.Error())
		} else {
			log.Printf("Subscribed to topic %s", topic)
		}
	})
	options.SetConnectionLostHandler(func(client paho.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	options.SetDefaultPublishHandler(func(client paho.Client, msg paho.Message) {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		payload := msg.Payload()
		var received Message

		if err := json.Unmarshal(payload, &received); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		if _, ok := reflect.TypeOf(received).FieldByName("Print"); ok {
			c.data = received
			log.Printf("Updated data: %v", c.data)
		}
	})

	c.client = paho.NewClient(options)
	return c
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

func (c *Client) Publish(command *Command) error {
	rawCommand, err := command.JSON()
	if err != nil {
		return err
	}

	token := c.client.Publish(fmt.Sprintf(commandTopic, c.config.Serial), 0, false, rawCommand)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) update() error {
	if !(time.Since(c.lastUpdate) > c.config.Timeout) {
		return errors.New("timeout")
	}

	c.lastUpdate = time.Now()
	// Return of this message is caught by the onmessage handler which update c.data
	return c.Publish(NewCommand(Pushing).AddCommandField("push_all"))
}

func (c *Client) Data() Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.update(); err != nil {
		return c.data
	}

	return c.data
}
