package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	clientID = "torbenconto/bambulabs_api"
)

type MqttConfig struct {
	Host         net.IP
	Port         int
	SerialNumber string
	AccessCode   string

	Timeout time.Duration
}

type MqttClient struct {
	ctx    context.Context
	cancel context.CancelFunc

	config      *MqttConfig
	client      paho.Client
	messageChan chan []byte
}

func (c *MqttClient) MessageChan() <-chan []byte {
	return c.messageChan
}

func NewMqttClient(parent context.Context, cfg *MqttConfig) (*MqttClient, error) {
	ctx, cancel := context.WithCancel(parent)

	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("mqtts://%s:%d", cfg.Host, cfg.Port)).
		SetClientID(clientID).
		SetUsername("bblp").
		SetPassword(cfg.AccessCode).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAutoReconnect(false).
		SetConnectTimeout(cfg.Timeout).
		SetWriteTimeout(cfg.Timeout).
		SetKeepAlive(30 * time.Second)

	client := &MqttClient{
		config:      cfg,
		client:      paho.NewClient(opts),
		ctx:         ctx,
		messageChan: make(chan []byte, 200),
		cancel:      cancel,
	}

	opts.SetOnConnectHandler(client.onConnect)

	return client, nil
}

func (c *MqttClient) onConnect(client paho.Client) {
	topic := fmt.Sprintf("device/%s/report", c.config.SerialNumber)
	token := client.Subscribe(topic, 0, c.handleMessage)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to subscribe to topic %s: %v", topic, token.Error())
		return
	}
	log.Printf("Subscribed to topic %s", topic)
}

func (c *MqttClient) handleMessage(_ paho.Client, msg paho.Message) {
	select {
	case c.messageChan <- msg.Payload():
	default:
		log.Println("Message dropped: channel full")
	}
}

func (m *MqttClient) Connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *MqttClient) Close() {
	c.cancel()
	close(c.messageChan)
	c.client.Disconnect(250)
}
