package bambulabs_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/ftp"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/internal/protocol"

	goftp "github.com/jlaffaye/ftp"
)

const defaultOpTimeout = 10 * time.Second

type Config struct {
	Host     net.IP
	MQTTPort int
	FTPPort  int
	Model    Model

	AccessCode   string
	SerialNumber string
}

type Printer interface {
	Serial() string
	Close() error
	State() (*mqtt.Message, bool)
	RequestUpdate() error

	SetLight(light Light, mode LightMode) error

	SetFan(fan Fan, speed uint8) error

	SendGcode(input []string) error

	ListFiles(path string) ([]*goftp.Entry, error)
	DownloadFile(path string, w io.Writer) error
	UploadFile(path string, r io.Reader) error
	DeleteFile(path string) error
}

type printer struct {
	serial string
	model  Model

	ctx    context.Context
	cancel context.CancelFunc
	mqtt   *mqtt.MqttClient
	ftp    *ftp.FtpClient

	state atomic.Pointer[mqtt.Message]

	done chan struct{}
}

func NewPrinter(parent context.Context, cfg Config) (*printer, error) {
	ctx, cancel := context.WithCancel(parent)

	mqttPort := cfg.MQTTPort
	if mqttPort == 0 {
		mqttPort = 8883
	}

	ftpPort := cfg.FTPPort
	if ftpPort == 0 {
		ftpPort = 990
	}

	mc, err := mqtt.NewMqttClient(&mqtt.MqttConfig{
		Host:         cfg.Host,
		Port:         mqttPort,
		SerialNumber: cfg.SerialNumber,
		Username:     "bblp",
		AccessCode:   cfg.AccessCode,
	})

	if err != nil {
		cancel()
		return nil, err
	}

	connectCtx, connectCancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer connectCancel()
	if err := mc.Connect(connectCtx); err != nil {
		cancel()
		return nil, err
	}

	ftpConnectCtx, ftpConnectCancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer ftpConnectCancel()
	fc, err := ftp.NewFtpClient(ftpConnectCtx, &ftp.FtpClientConfig{
		Host:       cfg.Host.String(),
		Port:       ftpPort,
		Username:   "bblp",
		AccessCode: cfg.AccessCode,
	})
	if err != nil {
		log.Printf("[%s] ftp connect failed, continuing without file access: %v", cfg.SerialNumber, err)
		fc = nil
	}

	p := &printer{
		serial: cfg.SerialNumber,
		model:  cfg.Model,

		ctx:    ctx,
		cancel: cancel,
		mqtt:   mc,
		ftp:    fc,
		done:   make(chan struct{}),
	}

	waitCtx, waitCancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer waitCancel()
	if err := p.waitForConnection(waitCtx); err != nil {
		cancel()
		_ = mc.Close()
		_ = fc.Close()
		return nil, err
	}

	p.run()

	return p, nil
}

func (p *printer) waitForConnection(ctx context.Context) error {
	return p.mqtt.WaitConnected(ctx)
}

func (p *printer) run() {
	messageChan := p.mqtt.MessageChan()

	go func() {
		defer close(p.done)

		for {
			select {
			case <-p.ctx.Done():
				return

			case <-p.mqtt.Done():
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

func (p *printer) publish(cmd *protocol.Command) error {
	ctx, cancel := context.WithTimeout(p.ctx, defaultOpTimeout)
	defer cancel()
	return p.mqtt.Publish(ctx, cmd)
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
	return p.publish(protocol.NewCommand(protocol.Pushing).WithCommand("pushall"))
}

func (p *printer) Serial() string {
	return p.serial
}

func (p *printer) Close() error {
	p.cancel()

	mqttErr := p.mqtt.Close()
	ftpErr := p.ftp.Close()

	<-p.done

	if mqttErr != nil {
		return mqttErr
	}
	return ftpErr
}

func (p *printer) State() (*mqtt.Message, bool) {
	m := p.state.Load()
	if m == nil {
		return nil, false
	}
	return m, true
}

// files (FTP)

func (p *printer) ListFiles(path string) ([]*goftp.Entry, error) {
	if p.ftp == nil {
		return nil, ErrFTPUnavailable
	}
	ctx, cancel := context.WithTimeout(p.ctx, defaultOpTimeout)
	defer cancel()
	return p.ftp.List(ctx, path)
}

func (p *printer) DownloadFile(path string, w io.Writer) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	ctx, cancel := context.WithTimeout(p.ctx, defaultOpTimeout)
	defer cancel()
	return p.ftp.Retrieve(ctx, path, w)
}

func (p *printer) UploadFile(path string, r io.Reader) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	ctx, cancel := context.WithTimeout(p.ctx, defaultOpTimeout)
	defer cancel()
	return p.ftp.Store(ctx, path, r)
}

func (p *printer) DeleteFile(path string) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	ctx, cancel := context.WithTimeout(p.ctx, defaultOpTimeout)
	defer cancel()
	return p.ftp.Delete(ctx, path)
}

// end files

// lights

func (p *printer) SetLight(light Light, mode LightMode) error {
	if !SupportsLight(p.model, light) {
		return fmt.Errorf("%w: %s", ErrLightNotSupported, light)
	}

	command := protocol.NewCommand(protocol.System).
		WithCommand("ledctrl").
		Set("led_node", light).
		Set("led_mode", mode).
		Set("led_on_time", 500).
		Set("led_off_time", 500).
		Set("loop_times", 1).
		Set("interval_time", 1000)

	if err := p.publish(command); err != nil {
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	return nil
}

// end lights

// begin fans

func (p *printer) SetFan(fan Fan, speed uint8) error { // implicit cap of 255
	if !SupportsFan(p.model, fan) {
		return fmt.Errorf("%w: %s", ErrFanNotSupported, fan.String())
	}

	if err := p.SendGcode([]string{fmt.Sprintf("M106 P%d S%d", fan, speed)}); err != nil {
		return fmt.Errorf("error setting fan %s: %w", fan, err)
	}
	return nil
}

// end fans

func (p *printer) SendGcode(input []string) error {
	for _, line := range input {
		// TODO: validate GCODE
		cmd := protocol.NewCommand(protocol.Print).WithCommand("gcode_line").WithParam(line)

		if err := p.publish(cmd); err != nil {
			return fmt.Errorf("failed to publish gcode line %s: %w", line, err)
		}
	}

	return nil
}
