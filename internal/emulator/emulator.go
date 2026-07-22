package emulator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type Emulator struct {
	cancel context.CancelFunc
	done   chan struct{}

	broker *mochi.Server

	serial string

	mu    sync.RWMutex
	state protocol.Report
}

func Start(ctx context.Context, cfg *bambulabs_api.Config, port int, reportFile string) (*Emulator, error) {
	ctx, cancel := context.WithCancel(ctx)

	raw, err := os.ReadFile(reportFile)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("read report: %w", err)
	}

	var initial protocol.Report
	if err := json.Unmarshal(raw, &initial); err != nil {
		cancel()
		return nil, fmt.Errorf("parse report: %w", err)
	}

	server := mochi.New(&mochi.Options{
		InlineClient: true,
	})

	if err := server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{
				{
					Username: "bblp",
					Password: auth.RString(cfg.AccessCode),
					Allow:    true,
				},
			},
			ACL: auth.ACLRules{
				{
					Username: "bblp",
					Filters: auth.Filters{
						"#": auth.ReadWrite,
					},
				},
			},
		},
	}); err != nil {
		cancel()
		return nil, err
	}

	tlsCfg, err := selfSignedTLS()
	if err != nil {
		cancel()
		return nil, err
	}

	if err := server.AddListener(listeners.NewTCP(listeners.Config{
		ID:        "emulator",
		Address:   fmt.Sprintf("127.0.0.1:%d", port),
		TLSConfig: tlsCfg,
	})); err != nil {
		cancel()
		return nil, err
	}

	go server.Serve()

	e := &Emulator{
		cancel: cancel,
		done:   make(chan struct{}),
		broker: server,
		serial: cfg.SerialNumber,
		state:  initial,
	}

	go e.run(ctx)

	return e, nil
}

func (e *Emulator) run(ctx context.Context) {
	defer close(e.done)

	_ = e.broker.Subscribe(
		fmt.Sprintf("device/%s/request", e.serial),
		1,
		e.handleRequest,
	)

	<-ctx.Done()

	_ = e.broker.Close()
}

func (e *Emulator) handleRequest(
	_ *mochi.Client,
	_ packets.Subscription,
	pk packets.Packet,
) {
	e.applyCommand(pk.Payload)
	e.publishReport()
}

// commandKey identifies a single (message type, command) pair, e.g. the
// "ledctrl" command sent under the "system" message type.
type commandKey struct {
	msgType string
	command string
}

// commandHandlers maps a recognized command onto the state mutation it
// causes. Add an entry here whenever the emulator needs to understand a new
// command (e.g. AMS filament changes, print state transitions).
var commandHandlers = map[commandKey]func(*Emulator, map[string]any){
	{msgType: "system", command: "ledctrl"}: (*Emulator).applyLedCtrl,
}

// applyCommand parses a raw MQTT request payload and mutates emulator state
// for any command it recognizes. Requests are grouped by message type
// (print/system/pushing/...), matching the shape produced by protocol.Command.
func (e *Emulator) applyCommand(payload []byte) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(payload, &envelope); err != nil {
		log.Printf("[emulator %s] failed to parse command: %v", e.serial, err)
		return
	}

	for msgType, raw := range envelope {
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			log.Printf("[emulator %s] failed to parse %q fields: %v", e.serial, msgType, err)
			continue
		}

		cmd, _ := fields["command"].(string)

		if handler, ok := commandHandlers[commandKey{msgType: msgType, command: cmd}]; ok {
			handler(e, fields)
		}
	}
}

func (e *Emulator) applyLedCtrl(fields map[string]any) {
	node, _ := fields["led_node"].(string)
	mode, _ := fields["led_mode"].(string)
	if node == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state.Print == nil {
		e.state.Print = &protocol.PrintReport{}
	}

	for i := range e.state.Print.LightsReport {
		if e.state.Print.LightsReport[i].Node == node {
			e.state.Print.LightsReport[i].Mode = mode
			return
		}
	}

	e.state.Print.LightsReport = append(e.state.Print.LightsReport, protocol.LightsReport{
		Node: node,
		Mode: mode,
	})
}

func (e *Emulator) publishReport() {
	e.mu.RLock()
	payload, err := json.Marshal(e.state)
	e.mu.RUnlock()
	if err != nil {
		log.Printf("[emulator %s] failed to marshal report: %v", e.serial, err)
		return
	}

	_ = e.broker.Publish(
		fmt.Sprintf("device/%s/report", e.serial),
		payload,
		false,
		0,
	)
}

// PushUpdate republishes the emulator's current state without waiting for a
// request from a client.
func (e *Emulator) PushUpdate() {
	e.publishReport()
}

// State returns a snapshot of the emulator's current report state. Intended
// for tests that want to assert on emulator-side truth directly, independent
// of whether a connected printer's decoder is behaving correctly.
func (e *Emulator) State() protocol.Report {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

func (e *Emulator) Stop() {
	e.cancel()
	<-e.done
}
