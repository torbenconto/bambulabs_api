package bambulabs_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/ftp"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Config struct {
	Host     net.IP
	MQTTPort int
	FTPPort  int
	Model    Model

	AccessCode   string
	SerialNumber string
}

const defaultOpTimeout time.Duration = 10 * time.Second

type Printer interface {
	Serial() string
	Close() error
	State() (*mqtt.Message, bool)

	RequestUpdate(ctx context.Context) error

	SetLight(ctx context.Context, light Light, mode LightMode) error
	SetFan(ctx context.Context, fan Fan, speed uint8) error
	SendGcode(ctx context.Context, input []string) error

	ListFiles(path string) ([]os.FileInfo, error)
	DownloadFile(path string, w io.Writer) error
	UploadFile(path string, r io.Reader) error
	DeleteFile(path string) error
}

type printer struct {
	serial string
	model  Model

	cancel context.CancelFunc

	mqtt *mqtt.MqttClient
	ftp  *ftp.FtpClient

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
	if err := mc.Connect(ctx); err != nil {
		cancel()
		return nil, err
	}

	fc := ftp.NewFtpClient(&ftp.FtpClientConfig{
		Host:       cfg.Host.String(),
		Port:       ftpPort,
		Username:   "bblp",
		AccessCode: cfg.AccessCode,
	})

	if err := fc.Connect(ctx); err != nil {
		log.Printf("[%s] ftp connect failed, continuing without file access: %v", cfg.SerialNumber, err)
		fc = nil
	}

	p := &printer{
		serial: cfg.SerialNumber,
		model:  cfg.Model,

		mqtt: mc,
		ftp:  fc,

		done:   make(chan struct{}),
		cancel: cancel,
	}

	if err := p.mqtt.WaitConnected(ctx); err != nil {
		_ = mc.Close()
		_ = fc.Close()
		return nil, err
	}

	p.run(ctx)

	return p, nil
}

func (p *printer) run(ctx context.Context) {
	messageChan := p.mqtt.MessageChan()

	go func() {
		defer close(p.done)

		for {
			select {
			case <-ctx.Done():
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

func (p *printer) publish(ctx context.Context, cmd *protocol.Command) error {
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
func (p *printer) RequestUpdate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer cancel()

	return p.publish(ctx, protocol.NewCommand(protocol.Pushing).WithCommand("pushall"))
}

func (p *printer) Serial() string {
	return p.serial
}

func (p *printer) Close() error {
	p.cancel()

	<-p.done

	mqttErr := p.mqtt.Close()

	var ftpErr error
	if p.ftp != nil {
		ftpErr = p.ftp.Close()
	}

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

func (p *printer) ListFiles(path string) ([]os.FileInfo, error) {
	if p.ftp == nil {
		return nil, ErrFTPUnavailable
	}

	return p.ftp.List(path)
}

func (p *printer) DownloadFile(path string, w io.Writer) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}

	return p.ftp.Retrieve(path, w)
}

func (p *printer) UploadFile(path string, r io.Reader) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	return p.ftp.Store(path, r)
}

func (p *printer) DeleteFile(path string) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	return p.ftp.Delete(path)
}

// end files

// lights

// SetLight
func (p *printer) SetLight(ctx context.Context, light Light, mode LightMode) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer cancel()

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

	if err := p.publish(ctx, command); err != nil {
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	return nil
}

// end lights

// begin fans

func (p *printer) SetFan(ctx context.Context, fan Fan, speed uint8) error { // implicit cap of 255
	ctx, cancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer cancel()

	if !SupportsFan(p.model, fan) {
		return fmt.Errorf("%w: %s", ErrFanNotSupported, fan.String())
	}

	if err := p.SendGcode(ctx, []string{fmt.Sprintf("M106 P%d S%d", fan, speed)}); err != nil {
		return fmt.Errorf("error setting fan %s: %w", fan, err)
	}
	return nil
}

// end fans

func (p *printer) SendGcode(ctx context.Context, input []string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOpTimeout)
	defer cancel()

	for _, line := range input {
		// TODO: validate GCODE
		cmd := protocol.NewCommand(protocol.Print).WithCommand("gcode_line").WithParam(line)

		if err := p.publish(ctx, cmd); err != nil {
			return fmt.Errorf("failed to publish gcode line %s: %w", line, err)
		}
	}

	return nil
}
