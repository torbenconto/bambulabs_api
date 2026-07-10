package mqtt

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
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
	Username     string
	SerialNumber string
	AccessCode   string
}

type MqttClient struct {
	config      *MqttConfig
	client      paho.Client
	messageChan chan []byte
	connected   chan struct{}

	closeOnce sync.Once
	stop      chan struct{}
}

func (c *MqttClient) MessageChan() <-chan []byte {
	return c.messageChan
}

func NewMqttClient(cfg *MqttConfig) (*MqttClient, error) {
	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("mqtts://%s:%d", cfg.Host.String(), cfg.Port)).
		SetClientID(clientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.AccessCode).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAutoReconnect(true).
		SetKeepAlive(30 * time.Second)

	client := &MqttClient{
		config:      cfg,
		messageChan: make(chan []byte, 200),
		stop:        make(chan struct{}),
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
	if err := waitToken(context.Background(), c.stop, token); err != nil {
		log.Printf("MQTT subscribe failed: %v", err)
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
	case <-c.stop:
		return
	case c.messageChan <- msg.Payload():
	default:
		// drop msg
	}
}

func (c *MqttClient) Connect(ctx context.Context) error {
	token := c.client.Connect()
	return waitToken(ctx, c.stop, token)
}

func (c *MqttClient) WaitConnected(ctx context.Context) error {
	select {
	case <-c.connected:
		return nil
	case <-c.stop:
		return errors.New("mqtt client closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *MqttClient) Connected() <-chan struct{} {
	return c.connected
}

func (c *MqttClient) Publish(ctx context.Context, cmd *protocol.Command) error {
	json, err := cmd.Marshal()
	if err != nil {
		return err
	}

	topic := fmt.Sprintf("device/%s/request", c.config.SerialNumber)

	token := c.client.Publish(topic, qos, false, json)
	return waitToken(ctx, c.stop, token)
}

func (c *MqttClient) Close() error {
	c.closeOnce.Do(func() {
		close(c.stop)
		c.client.Disconnect(250)
	})
	return nil
}

func (c *MqttClient) Done() <-chan struct{} {
	return c.stop
}

func waitToken(ctx context.Context, stopChan <-chan struct{}, t paho.Token) error {
	select {
	case <-t.Done():
		return t.Error()
	case <-stopChan:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}
