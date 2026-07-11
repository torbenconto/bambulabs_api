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
	State() (*mqtt.Message, bool)

	RequestUpdate(ctx context.Context) error

	SetLight(ctx context.Context, light Light, mode LightMode) error
	SetLightFlashing(ctx context.Context, light Light, cfg LightFlashingConfig) error
	SetFan(ctx context.Context, fan Fan, speed uint8) error
	SendGcode(ctx context.Context, input []string) error

	ListFiles(path string) ([]os.FileInfo, error)
	DownloadFile(path string, w io.Writer) error
	UploadFile(path string, r io.Reader) error
	DeleteFile(path string) error
}

type printer struct {
	cfg Config // own the config

	// Cancellation tree for entire printer object
	cancel context.CancelFunc

	mqtt *mqtt.MqttClient
	ftp  *ftp.FtpClient

	// Hot-swappable pointer to the current mqtt state
	// May represent some leakage of information but neccessary in order to simply state access mechanisms
	state atomic.Pointer[mqtt.Message]

	done chan struct{}
}

// NewPrinter creates a new [printer] object and attempts both an MQTT and FTP connection using provided options
// If the MQTT connection fails, the construction fails. If the FTP fails, construction will succeed but remain in a degraded state.
func NewPrinter(parent context.Context, cfg Config) (*printer, error) {
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

	fc := ftp.NewFtpClient(&ftp.FtpClientConfig{
		Host:       cfg.Host.String(),
		Port:       ftpPort,
		Username:   "bblp",
		AccessCode: cfg.AccessCode,
	})

	// FTP is non-vital so we'll warn the user and proceed without FTP connection.
	if err := fc.Connect(ctx); err != nil {
		log.Printf("[%s] ftp connect failed, continuing without file access: %v", cfg.SerialNumber, err)
		fc = nil
	}

	p := &printer{
		cfg: cfg,

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

	// run state loop (goroutine)
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

// command publishing helper, possibly include some checks in the future
func (p *printer) publish(ctx context.Context, cmd *protocol.Command) error {
	return p.mqtt.Publish(ctx, cmd)
}

// updateState takes a raw MQTT payload and attempts to convert it into a [import/mqtt.Message].
// Failure is not fatal but may represent something severly wrong with the message struct itself.
func (p *printer) updateState(payload []byte) {
	var msg mqtt.Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[%s] failed to unmarshal MQTT payload: %v", p.cfg.SerialNumber, err)
		return
	}

	p.state.Store(&msg)
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

// State returns the current MQTT state as a [import/mqtt.Message] alongside a boolean indicating a successful retrieve
func (p *printer) State() (*mqtt.Message, bool) {
	m := p.state.Load()
	if m == nil {
		return nil, false
	}
	return m, true
}

// files (FTP)

// ListFiles calls the underlying FTP client to fetch files found on the printer, returns an [ErrFTPUnavalible] if FTP is unavalible.
func (p *printer) ListFiles(path string) ([]os.FileInfo, error) {
	if p.ftp == nil {
		return nil, ErrFTPUnavailable
	}

	return p.ftp.List(path)
}

// DownloadFile calls the underlying FTP client to retrieve a file found on the printer to an [import/io.Writer], returns an [ErrFTPUnavalible] if FTP is unavalible.
func (p *printer) DownloadFile(path string, w io.Writer) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}

	return p.ftp.Retrieve(path, w)
}

// UploadFile calls the underlying FTP client to upload a file (given as an [import/io.Reader]) to a given path, returns an [ErrFTPUnavalible] if FTP is unavalible.
func (p *printer) UploadFile(path string, r io.Reader) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	return p.ftp.Store(path, r)
}

// DeleteFile calls the underlying FTP client to delete a file off of the printer (by path), returns an [ErrFTPUnavalible] if FTP is unavalible.
func (p *printer) DeleteFile(path string) error {
	if p.ftp == nil {
		return ErrFTPUnavailable
	}
	return p.ftp.Delete(path)
}

// end files

// lights

// SetLight publishes an MQTT command to control a given [Light], allowing you to set it to a given [LightMode].
// For [LightFlashing], [DefaultLightFlashingConfig] is used. Call [Printer.SetLightFlashing] to customize the flashing timing.
// If the [Printer] you attempt to call this function on does not support the chosen light, an [ErrLightNotSupported] will be returned.
func (p *printer) SetLight(ctx context.Context, light Light, mode LightMode) error {
	return p.setLight(ctx, light, mode, DefaultLightFlashingConfig())
}

// SetLightFlashing publishes an MQTT command that flashes a given [Light] using cfg.
// If the [Printer] does not support the chosen light, an [ErrLightNotSupported] will be returned.
func (p *printer) SetLightFlashing(ctx context.Context, light Light, cfg LightFlashingConfig) error {
	return p.setLight(ctx, light, LightFlashing, cfg)
}

func (p *printer) setLight(ctx context.Context, light Light, mode LightMode, cfg LightFlashingConfig) error {
	ctx, cancel := withDefaultOpTimeout(ctx)
	defer cancel()

	if !SupportsLight(p.cfg.Model, light) {
		return fmt.Errorf("%w: %s", ErrLightNotSupported, light)
	}

	command := newLightCommand(light, mode, cfg)

	if err := p.publish(ctx, command); err != nil {
		return fmt.Errorf("error setting light %s: %w", light, err)
	}

	return nil
}

func newLightCommand(light Light, mode LightMode, cfg LightFlashingConfig) *protocol.Command {
	return protocol.NewCommand(protocol.System).
		WithCommand("ledctrl").
		Set("led_node", light).
		Set("led_mode", mode).
		Set("led_on_time", cfg.OnTime.Milliseconds()).
		Set("led_off_time", cfg.OffTime.Milliseconds()).
		Set("loop_times", cfg.LoopTimes).
		Set("interval_time", cfg.IntervalTime.Milliseconds())
}

// end lights

// begin fans

// SetFan publishes a GCODE command (M106) via MQTT, allowing you to set a given [Fan] to a speed between 0-255.
// If the [Printer] you attempt to call this on does not support the chosen fan, an [ErrFanNotSupported] will be returned.
func (p *printer) SetFan(ctx context.Context, fan Fan, speed uint8) error { // implicit cap of 255
	ctx, cancel := withDefaultOpTimeout(ctx)
	defer cancel()

	if !SupportsFan(p.cfg.Model, fan) {
		return fmt.Errorf("%w: %s", ErrFanNotSupported, fan.String())
	}

	if err := p.SendGcode(ctx, []string{fmt.Sprintf("M106 P%d S%d", fan, speed)}); err != nil {
		return fmt.Errorf("error setting fan %s: %w", fan, err)
	}
	return nil
}

// end fans

// SendGcode sends raw GCODE commands to the printer via MQTT, be careful of what you send because the commands are currently not validated.
// EXERCISE CAUTION WHEN USING THIS FUNCTION, IT CAN AND WILL DAMAGE YOUR PRINTER IF USED IMPROPERLY
func (p *printer) SendGcode(ctx context.Context, input []string) error {
	ctx, cancel := withDefaultOpTimeout(ctx)
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
