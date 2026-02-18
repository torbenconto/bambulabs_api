package bambulabs_api

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/mqtt"
)

type Config struct {
	Host  net.IP
	Model Model

	AccessCode   string
	SerialNumber string
}

type Printer interface {
	Serial() string
	Close() error
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

	mc, err := mqtt.NewMqttClient(ctx, &mqtt.MqttConfig{
		Host:         cfg.Host,
		Port:         8883,
		SerialNumber: cfg.SerialNumber,
		AccessCode:   cfg.AccessCode,
		Timeout:      10 * time.Second,
	})

	if err != nil {
		cancel()
		return nil, err
	}

	if err := mc.Connect(); err != nil {
		return nil, err
	}

	p := &printer{
		serial: cfg.SerialNumber,
		model:  cfg.Model,

		ctx:    ctx,
		cancel: cancel,
		mqtt:   mc,
	}

	p.run()

	return p, nil
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

func (p *printer) Serial() string {
	return p.serial
}

func (p *printer) Close() error {
	p.cancel()
	err := p.mqtt.Close()

	<-p.done

	return err
}
