package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	clientID          = "golang-bambulabs-api"
	topicTemplate     = "device/%s/report"
	commandTopic      = "device/%s/request"
	qos               = 0
	connectionTimeout = 10 * time.Second
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
	messageChan chan []byte
	doneChan    chan struct{}
}

// NewClient creates and initializes a new MQTT client.
func NewClient(config *ClientConfig) *Client {
	options := paho.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtts://%s:%d", config.Host, config.Port))
	options.SetClientID(clientID)
	options.SetUsername(config.Username)
	options.SetPassword(config.AccessCode)
	options.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	options.SetAutoReconnect(true)

	client := &Client{
		config:      config,
		data:        Message{},
		lastUpdate:  time.Now(),
		messageChan: make(chan []byte, 200),
		doneChan:    make(chan struct{}),
	}

	options.SetOnConnectHandler(client.handleConnect)
	options.SetConnectionLostHandler(client.handleConnectionLost)
	options.SetDefaultPublishHandler(client.handleMessage)

	client.client = paho.NewClient(options)

	go client.processMessages()
	go client.periodicUpdate()

	return client
}

// Connect establishes a connection to the MQTT broker.
func (c *Client) Connect() error {
	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	log.Println("MQTT client connected")
	return nil
}

// Disconnect gracefully disconnects from the MQTT broker.
func (c *Client) Disconnect() {
	close(c.doneChan)
	c.client.Disconnect(uint(connectionTimeout.Milliseconds()))
	log.Println("MQTT client disconnected")
}

// Publish sends a command message to the MQTT broker.
func (c *Client) Publish(command *Command) error {
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

// Data retrieves the latest data, updating if necessary.
func (c *Client) Data() Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if time.Since(c.lastUpdate) > c.config.Timeout {
		if err := c.update(); err != nil {
			log.Printf("Failed to update data: %v", err)
		}
	}
	return c.data
}

// update refreshes the client data by pushing a "push_all" command.
func (c *Client) update() error {
	if time.Since(c.lastUpdate) <= c.config.Timeout {
		return errors.New("update called before timeout")
	}
	c.lastUpdate = time.Now()

	command := NewCommand(Pushing).AddCommandField("push_all")
	return c.Publish(command)
}

// handleConnect subscribes to the required topic upon successful connection.
func (c *Client) handleConnect(client paho.Client) {
	topic := fmt.Sprintf(topicTemplate, c.config.Serial)
	token := client.Subscribe(topic, qos, nil)
	if token.Wait() && token.Error() != nil {
		log.Printf("Error subscribing to topic %s: %v", topic, token.Error())
	} else {
		log.Printf("Subscribed to topic %s", topic)
	}
}

// handleConnectionLost logs the connection loss.
func (c *Client) handleConnectionLost(client paho.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

// handleMessage enqueues incoming messages for processing.
func (c *Client) handleMessage(client paho.Client, msg paho.Message) {
	select {
	case c.messageChan <- msg.Payload():
		log.Printf("Message enqueued: %s", msg.Topic())
	default:
		log.Println("Message channel full; dropping message")
	}
}

// processMessages processes messages from the channel and updates the data.
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

// processPayload validates and updates the client data.
func (c *Client) processPayload(payload []byte) {
	var received Message
	if err := json.Unmarshal(payload, &received); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Perform fine-grained updates
	mergeMessages(&c.data, &received)
	log.Printf("Updated data: %v", c.data)
}

// mergeMessages updates fields in the existing message with those from the new message.
func mergeMessages(existing, new *Message) {
	existingValue := reflect.ValueOf(existing).Elem()
	newValue := reflect.ValueOf(new).Elem()

	for i := 0; i < existingValue.NumField(); i++ {
		field := existingValue.Type().Field(i)
		existingField := existingValue.Field(i)
		newField := newValue.Field(i)

		// Only update if the new field is valid and has a non-zero value
		if newField.IsValid() && !isZeroValue(newField) {
			if existingField.Kind() == reflect.Struct {
				// Recursively merge nested structs
				mergeMessages(existingField.Addr().Interface().(*Message), newField.Addr().Interface().(*Message))
			} else {
				existingField.Set(newField)
				log.Printf("Updated field '%s': %v", field.Name, newField.Interface())
			}
		}
	}
}

func (c *Client) periodicUpdate() {
	ticker := time.NewTicker(c.config.Timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.update(); err != nil {
				log.Printf("Failed to update data: %v", err)
			}
		case <-c.doneChan:
			return
		}
	}
}

// isZeroValue checks if a field is its zero value (e.g., 0 for int, "" for string).
func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
