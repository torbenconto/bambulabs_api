package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	clientID       = "golang-bambulabs-api"
	topicTemplate  = "device/%s/report"
	commandTopic   = "device/%s/request"
	qos            = 0
	updateInterval = 10 * time.Second
)

// ClientConfig holds the configuration details for the MQTT client.
type ClientConfig struct {
	Host       string
	Port       int
	Serial     string
	Username   string
	AccessCode string
	Timeout    time.Duration
}

// Client represents the MQTT client.
type Client struct {
	config      *ClientConfig
	client      paho.Client
	mutex       sync.Mutex
	data        Message
	lastUpdate  time.Time
	sequenceID  int
	messageChan chan []byte
	doneChan    chan struct{}
	ticker      *time.Ticker
}

// NewClient initializes a new MQTT client.
func NewClient(config *ClientConfig) *Client {
	// Need to add an option to set OrderMatters for this
	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("mqtts://%s:%d", config.Host, config.Port)).
		SetClientID(clientID).
		SetUsername(config.Username).
		SetPassword(config.AccessCode).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAutoReconnect(true).
		SetKeepAlive(30 * time.Second)

	client := &Client{
		config:      config,
		messageChan: make(chan []byte, 200),
		doneChan:    make(chan struct{}),
		sequenceID:  0,
		ticker:      time.NewTicker(updateInterval),
	}

	opts.SetOnConnectHandler(client.onConnect)
	opts.SetConnectionLostHandler(client.onConnectionLost)
	opts.SetDefaultPublishHandler(client.handleMessage)

	client.client = paho.NewClient(opts)

	return client
}

// Connect establishes a connection to the MQTT broker.
func (c *Client) Connect() error {
	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	log.Println("Connected to MQTT broker")
	go c.processMessages()
	go c.periodicUpdate()
	return nil
}

// Disconnect gracefully closes the connection.
func (c *Client) Disconnect() {
	close(c.doneChan)
	c.ticker.Stop()
	c.client.Disconnect(250)
	log.Println("Disconnected from MQTT broker")
}

// Publish sends a command message to the MQTT broker.
func (c *Client) Publish(command *Command) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.sequenceID++
	command.SetId(c.sequenceID)

	rawCommand, err := command.JSON()
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	topic := fmt.Sprintf(commandTopic, c.config.Serial)
	token := c.client.Publish(topic, qos, false, rawCommand)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, token.Error())
	}

	log.Printf("Published command to topic %s", topic)
	return nil
}

// Data retrieves the latest data and triggers an update if stale.
func (c *Client) Data() Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if time.Since(c.lastUpdate) > c.config.Timeout {
		go c.update()
	}
	return c.data
}

// Private methods

// update triggers a data refresh by publishing a "push_all" command.
func (c *Client) update() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if time.Since(c.lastUpdate) < c.config.Timeout {
		return
	}

	c.lastUpdate = time.Now()
	command := NewCommand(Pushing).AddCommandField("pushall")
	js, _ := command.JSON()
	fmt.Println("command", js)
	if err := c.Publish(command); err != nil {
		log.Printf("Failed to publish update command: %v", err)
	}
}

func (c *Client) periodicUpdate() {
	for {
		select {
		case <-c.ticker.C:
			c.update()
		case <-c.doneChan:
			return
		}
	}
}

// onConnect subscribes to the data topic.
func (c *Client) onConnect(client paho.Client) {
	topic := fmt.Sprintf(topicTemplate, c.config.Serial)
	token := client.Subscribe(topic, qos, nil)
	if token.Wait() && token.Error() != nil {
		log.Printf("Failed to subscribe to topic %s: %v", topic, token.Error())
		return
	}
	log.Printf("Subscribed to topic %s", topic)
}

// onConnectionLost logs connection loss.
func (c *Client) onConnectionLost(client paho.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

// handleMessage queues incoming messages for processing.
func (c *Client) handleMessage(client paho.Client, msg paho.Message) {
	select {
	case c.messageChan <- msg.Payload():
		log.Printf("Message received: %s", msg.Topic())
	default:
		log.Println("Message dropped: channel full")
	}
}

// processMessages processes incoming messages from the channel.
func (c *Client) processMessages() {
	for {
		select {
		case payload := <-c.messageChan:
			c.processPayload(payload)
		case <-c.doneChan:
			return
		}
	}
}

// processPayload updates the client data with the incoming message.
func (c *Client) processPayload(payload []byte) {
	var received Message
	if err := json.Unmarshal(payload, &received); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	mergeMessages(&c.data, &received)
	//log.Printf("Updated data: %+v", c.data)
}

// mergeMessages recursively merges the existing and new messages.
func mergeMessages(existing, new *Message) {
	// Use reflection to iterate through the fields of the "Print" struct.
	mergeStructs(&existing.Print, &new.Print)
}

// mergeStructs dynamically merges fields of two structs using reflection.
func mergeStructs(existing, new interface{}) {
	existingVal := reflect.ValueOf(existing).Elem()
	newVal := reflect.ValueOf(new).Elem()

	// Iterate over each field in the struct.
	for i := 0; i < existingVal.NumField(); i++ {
		field := existingVal.Field(i)

		// Ensure that the field is a valid field to merge.
		newField := newVal.Field(i)
		if !newField.IsValid() {
			continue
		}

		// Only merge if the field is non-zero in the new struct.
		if !newField.IsZero() {
			// If it's a struct, recursively merge it.
			if newField.Kind() == reflect.Struct {
				mergeStructs(field.Addr().Interface(), newField.Addr().Interface())
			} else {
				// Otherwise, set the field to the new value.
				field.Set(newField)
			}
		}
	}
}
