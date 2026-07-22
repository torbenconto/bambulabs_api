package bambulabs_api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/torbenconto/bambulabs_api/internal/ftp"
	"github.com/torbenconto/bambulabs_api/internal/mqtt"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

// defaultOpTimeout is applied when a caller does not provide a deadline.
const defaultOpTimeout time.Duration = 10 * time.Second

func withDefaultOpTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, defaultOpTimeout)
}

// Config represents configuration options for a given [Printer], changing the MQTT and FTP ports is not recommended for inexperienced users (mostly used for testing purposes with the emulator).
type Config struct {
	Host     net.IP
	MQTTPort int
	FTPPort  int
	Model    Model

	AccessCode   string
	SerialNumber string
}

// Printer represents a connection to any and all BambuLabs printers, the primary [Client] struct holds objects that satisfy this interface.
type Printer interface {
	Serial() string
	Close() error
}

type printer struct {
	cfg Config

	commandClient CommandClient
	fileClient    FileClient

	mqtt *mqtt.MqttClient
	ftp  *ftp.FtpClient

	AMS *AMSSystem
	// Extruders *ExtruderSystem
	// Nozzles   *NozzleSystem
	Lights *LightSystem
	// Fans      *FanSystem
	// Files     *FileSystem

	cap Capability

	decoder Decoder

	mu   sync.RWMutex
	done chan struct{}

	cancel context.CancelFunc

	readyOnce sync.Once
	ready     chan struct{}
}

// NewPrinter creates a new [printer] object and attempts both an MQTT and FTP connection using provided options
// If the MQTT connection fails, the construction fails. If the FTP fails, construction will succeed but remain in a degraded state.
func NewPrinter(parent context.Context, cfg *Config) (*printer, error) {
	ctx, cancel := context.WithCancel(parent)

	// Assign default ports if none provided.
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

	// MQTT connection is vital for printer communication so we'll deconstruct the entire object if it fails.
	if err != nil {
		cancel()
		return nil, err
	}
	if err := mc.Connect(ctx); err != nil {
		cancel()
		return nil, err
	}

	commandClient := newMqttCommandClient(mc)

	fc := ftp.NewFtpClient(&ftp.FtpClientConfig{
		Host:       cfg.Host.String(),
		Port:       ftpPort,
		Username:   "bblp",
		AccessCode: cfg.AccessCode,
	})

	var (
		ftpClient  *ftp.FtpClient
		fileClient FileClient
	)

	// FTP is non-vital so we'll warn the user and proceed without file access.
	if err := fc.Connect(ctx); err != nil {
		log.Printf("[%s] ftp connect failed, continuing without file access: %v", cfg.SerialNumber, err)
	} else {
		ftpClient = fc
		fileClient = newFTPFileClient(fc)
	}

	p := &printer{
		cfg: *cfg,

		mqtt: mc,
		ftp:  ftpClient,

		commandClient: commandClient,
		fileClient:    fileClient,

		done:   make(chan struct{}),
		cancel: cancel,

		Lights: NewLightSystem(commandClient),
		AMS:    NewAMSSystem(),

		decoder: *NewDecoder(cfg.Model),
		ready:   make(chan struct{}),
	}

	if err := p.mqtt.WaitConnected(ctx); err != nil {
		_ = mc.Close()
		if ftpClient != nil {
			_ = ftpClient.Close()
		}
		return nil, err
	}
	// run state loop (goroutine)
	p.run(ctx)

	if err := p.awaitInitialState(ctx); err != nil {
		cancel() // stop run() goroutine
		<-p.done
		_ = mc.Close()
		if ftpClient != nil {
			_ = ftpClient.Close()
		}
		return nil, err
	}

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

func (p *printer) awaitInitialState(ctx context.Context) error {
	if err := p.RequestUpdate(ctx); err != nil {
		return fmt.Errorf("request initial state: %w", err)
	}

	select {
	case <-p.ready:
		return nil
	case <-p.mqtt.Done():
		return errors.New("mqtt connection closed before initial state was received")
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for initial printer state: %w", ctx.Err())
	}
}

// command publishing helper, possibly include some checks in the future
func (p *printer) publish(ctx context.Context, cmd *protocol.Command) error {
	return p.mqtt.Publish(ctx, cmd)
}

// updateState takes a raw MQTT payload and attempts to convert it into a [import/mqtt.Message].
// Failure is not fatal but may represent something severly wrong with the message struct itself.
func (p *printer) updateState(payload []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var report protocol.Report
	if err := json.Unmarshal(payload, &report); err != nil {
		log.Printf("[%s] failed to unmarshal MQTT payload: %v", p.cfg.SerialNumber, err)
		return
	}

	p.decoder.Apply(p, &report)

	p.readyOnce.Do(func() { close(p.ready) })
}

// RequestUpdate manually requests a "pushall", updating the printer state. Exercise caution in the interval you use this, especially on lower end printers.
func (p *printer) RequestUpdate(ctx context.Context) error {
	ctx, cancel := withDefaultOpTimeout(ctx)
	defer cancel()

	return p.publish(ctx, protocol.NewCommand(protocol.Pushing).WithCommand("pushall"))
}

// Serial returns the printer serial number provided during construction.
func (p *printer) Serial() string {
	return p.cfg.SerialNumber
}

// Close terminates the connection to the printer and it's underlying clients.
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
