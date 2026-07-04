package mqtt

import (
	"context"
	"crypto/tls"
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
	ctx    context.Context
	cancel context.CancelFunc

	config      *MqttConfig
	client      paho.Client
	messageChan chan []byte
	connected   chan struct{}

	closeOnce sync.Once
}

func (c *MqttClient) MessageChan() <-chan []byte {
	return c.messageChan
}

func NewMqttClient(parent context.Context, cfg *MqttConfig) (*MqttClient, error) {
	ctx, cancel := context.WithCancel(parent)

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
		ctx:         ctx,
		messageChan: make(chan []byte, 200),
		cancel:      cancel,
		connected:   make(chan struct{}),
	}

	opts.SetOnConnectHandler(client.onConnect)
	opts.SetDefaultPublishHandler(client.handleMessage)

	opts.SetConnectionLostHandler(func(c paho.Client, err error) {
		<-ctx.Done()
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
	case <-tokenDone(token):
		if token.Error() != nil {
			log.Printf("MQTT subscribe failed: %v", token.Error())
			return
		}
	case <-c.ctx.Done():
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
	case <-c.ctx.Done():
		return
	case c.messageChan <- msg.Payload():
	default:
		// drop msg
	}
}

func (c *MqttClient) Connect(ctx context.Context) error {
	token := c.client.Connect()
	select {
	case <-tokenDone(token):
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}

func (c *MqttClient) WaitConnected(ctx context.Context) error {
	select {
	case <-c.connected:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return c.ctx.Err()
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

	select {
	case <-tokenDone(token):
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}

func (c *MqttClient) Close() error {
	c.closeOnce.Do(func() {
		c.cancel()
		c.client.Disconnect(250)
	})
	return nil
}

func (c *MqttClient) Done() <-chan struct{} {
	return c.ctx.Done()
}

func tokenDone(t paho.Token) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		t.Wait()
		close(done)
	}()
	return done
}
