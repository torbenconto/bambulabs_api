package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

const (
	clientID = "torbenconto/bambulabs_api"
	qos      = 0
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
	connected   chan struct{}
}

func (c *MqttClient) MessageChan() <-chan []byte {
	return c.messageChan
}

func NewMqttClient(parent context.Context, cfg *MqttConfig) (*MqttClient, error) {
	ctx, cancel := context.WithCancel(parent)

	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("mqtts://%s:%d", cfg.Host.String(), cfg.Port)).
		SetClientID(clientID).
		SetUsername("bblp").
		SetPassword(cfg.AccessCode).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAutoReconnect(true).
		SetConnectTimeout(cfg.Timeout).
		SetWriteTimeout(cfg.Timeout).
		SetKeepAlive(30 * time.Second)

	client := &MqttClient{
		config:      cfg,
		ctx:         ctx,
		messageChan: make(chan []byte, 200),
		cancel:      cancel,
		connected:   make(chan struct{}),
	}

	opts.SetOnConnectHandler(client.onConnect)
	opts.SetDefaultPublishHandler(client.handleMessage)

	opts.SetConnectionLostHandler(func(c paho.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	client.client = paho.NewClient(opts)

	return client, nil
}

func (c *MqttClient) onConnect(client paho.Client) {
	topic := fmt.Sprintf("device/%s/report", c.config.SerialNumber)
	token := client.Subscribe(topic, 0, c.handleMessage)
	token.Wait()
	if token.Error() != nil {
		return
	}
	select {
	case <-c.connected:
	default:
		close(c.connected)
	}
}

func (c *MqttClient) handleMessage(_ paho.Client, msg paho.Message) {
	select {
	case c.messageChan <- msg.Payload():
	default:
		// drop msg
	}
}

func (c *MqttClient) Connect() error {
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *MqttClient) Connected() <-chan struct{} {
	return c.connected
}

func (c *MqttClient) Publish(cmd *protocol.Command) error {
	json, err := cmd.Marshal()
	if err != nil {
		return err
	}

	topic := fmt.Sprintf("device/%s/request", c.config.SerialNumber)

	if token := c.client.Publish(topic, qos, false, json); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *MqttClient) Close() error {
	c.cancel()
	close(c.messageChan)
	c.client.Disconnect(250)

	return nil
}
