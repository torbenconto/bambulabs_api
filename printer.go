package bambulabs_api

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Config struct {
	Host  net.IP
	Port  int
	Model Model

	AccessCode   string
	SerialNumber string
}

type Printer interface {
	Serial() string
	Close() error
	State() (*mqtt.Message, bool)
	RequestUpdate() error
}

type printer struct {
	serial string
	model  Model

	ctx    context.Context
	cancel context.CancelFunc
	mqtt   *mqtt.MqttClient

	state atomic.Pointer[mqtt.Message]

	done chan struct{}
}

func NewPrinter(parent context.Context, cfg Config) (*printer, error) {
	ctx, cancel := context.WithCancel(parent)

	port := cfg.Port
	if port == 0 {
		port = 8883
	}

	mc, err := mqtt.NewMqttClient(ctx, &mqtt.MqttConfig{
		Host:         cfg.Host,
		Port:         port,
		SerialNumber: cfg.SerialNumber,
		AccessCode:   cfg.AccessCode,
		Timeout:      10 * time.Second,
	})

	if err != nil {
		cancel()
		return nil, err
	}

	if err := mc.Connect(); err != nil {
		cancel()
		return nil, err
	}

	p := &printer{
		serial: cfg.SerialNumber,
		model:  cfg.Model,

		ctx:    ctx,
		cancel: cancel,
		mqtt:   mc,
		done:   make(chan struct{}),
	}

	p.run()

	return p, nil
}

func (p *printer) waitForConnection(ctx context.Context) error {
	select {
	case <-p.mqtt.Connected():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *printer) run() {
	messageChan := p.mqtt.MessageChan()

	go func() {
		defer close(p.done)

		for {
			select {
			case <-p.ctx.Done():
				return
			case payload, ok := <-messageChan:
				if !ok {
					return
				}
				p.updateState(payload)
			}
		}
	}()
}

func (p *printer) updateState(payload []byte) {
	var msg mqtt.Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[%s] failed to unmarshal MQTT payload: %v", p.serial, err)
		return
	}

	p.state.Store(&msg)
}

// RequestUpdate manually requests a "pushall", updating the printer state. Exercise caution in the interval you use this, especially on lower end printers.
func (p *printer) RequestUpdate() error {
	return p.mqtt.Publish(protocol.NewCommand(protocol.Pushing).WithCommand("pushall"))
}

func (p *printer) Serial() string {
	return p.serial
}

func (p *printer) Close() error {
	p.cancel()
	err := p.mqtt.Close()

	<-p.done

	return err
}

func (p *printer) State() (*mqtt.Message, bool) {
	m := p.state.Load()
	if m == nil {
		return nil, false
	}
	return m, true
}
