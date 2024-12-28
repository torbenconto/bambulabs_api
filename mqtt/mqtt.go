package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/torbenconto/bambulabs_api/types"
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
	Clientid = "golang-bambulabs-api"
	Topic    = "/device/%s/report"
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
	data       types.Data
	lastUpdate time.Time
}

func NewClient(config *ClientConfig) *Client {
	options := paho.NewClientOptions()
	// Maybe tls:// or tcp://
	options.AddBroker("mqtts://" + config.Host.String() + ":" + strconv.Itoa(config.Port))
	options.SetClientID(Clientid)
	options.SetUsername(config.Username)
	options.SetPassword(config.AccessCode)
	options.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	options.SetAutoReconnect(true)

	c := &Client{
		config:     config,
		data:       types.Data{},
		lastUpdate: time.Now(),
	}

	// TODO: needs improvement
	options.SetOnConnectHandler(func(client paho.Client) {
		topic := fmt.Sprintf(Topic, config.Serial)
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
		var received types.Data

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

func (c *Client) Publish(payload string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := c.client.Publish(Topic, 0, false, data)
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
	return c.Publish(Command{
		Type:    Pushing,
		Command: "push_all",
	}.JSON())
}

func (c *Client) Data() types.Data {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.update(); err != nil {
		return c.data
	}

	return c.data
}
