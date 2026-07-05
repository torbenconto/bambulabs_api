package emulator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type incomingCommand struct {
	Command    string `json:"command"`
	SequenceID string `json:"sequence_id"`
	Param      string `json:"param,omitempty"`
}

type Emulator struct {
	cancel                  context.CancelFunc
	port                    int
	host                    string
	broker                  *mochi.Server
	done                    chan struct{}
	targetModel             bambulabs_api.Model
	serial                  string
	capability              bambulabs_api.Capability
	gcodeState              bambulabs_api.GcodeState
	unsolicitedUpdateTicker *time.Ticker
	mu                      sync.Mutex
}

func Start(ctx context.Context, cfg *bambulabs_api.Config, port int) (*Emulator, error) {
	ctx, cancel := context.WithCancel(ctx)
	server := mochi.New(&mochi.Options{
		InlineClient: true,
	})
	if err := server.AddHook(new(auth.Hook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{
				{Username: "bblp", Password: auth.RString(cfg.AccessCode), Allow: true},
			},
			ACL: auth.ACLRules{
				{Username: "bblp", Filters: auth.Filters{
					"#": auth.ReadWrite,
				}},
			},
		},
	}); err != nil {
		cancel()
		return nil, fmt.Errorf("add auth hook %v", err)
	}

	tlsCfg, err := selfSignedTLS()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("generate tls cert: %v", err)
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:        "torbenconto/bambulabs_api/emulator",
		Address:   fmt.Sprintf("127.0.0.1:%d", port),
		TLSConfig: tlsCfg,
	})
	if err := server.AddListener(tcp); err != nil {
		cancel()
		return nil, fmt.Errorf("add tcp listener %v", err)
	}

	go func() {
		if err := server.Serve(); err != nil {
			return
		}
	}()

	emu := &Emulator{
		host:        "127.0.0.1",
		port:        port,
		broker:      server,
		cancel:      cancel,
		done:        make(chan struct{}),
		targetModel: cfg.Model,
		capability:  bambulabs_api.CapabilityAnyAms,
		serial:      cfg.SerialNumber,
		gcodeState:  bambulabs_api.IDLE,
	}

	emu.setTickers()
	go emu.run(ctx)
	return emu, nil
}

func (e *Emulator) setTickers() {
	switch e.targetModel {
	case bambulabs_api.ModelX1C, bambulabs_api.ModelX1E, bambulabs_api.ModelH2:
		e.unsolicitedUpdateTicker = time.NewTicker(5 * time.Second)
	default:
		e.unsolicitedUpdateTicker = time.NewTicker(1 * time.Minute)
	}
}

func (e *Emulator) run(ctx context.Context) {
	defer close(e.done)
	e.broker.Subscribe(fmt.Sprintf("device/%s/request", e.serial), 1, e.handleCommand)
	for {
		select {
		case <-ctx.Done():
			_ = e.broker.Close()
			return
		case <-e.unsolicitedUpdateTicker.C:
			e.publishCurrentState()
		}
	}
}

func (e *Emulator) handleCommand(_ *mochi.Client, _ packets.Subscription, pk packets.Packet) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(pk.Payload, &envelope); err != nil {
		return
	}
	for t, inner := range envelope {
		var cmd incomingCommand
		if err := json.Unmarshal(inner, &cmd); err != nil {
			continue
		}
		e.dispatch(protocol.MessageType(t), cmd)
	}
}

func (e *Emulator) dispatch(t protocol.MessageType, cmd incomingCommand) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch t {
	case protocol.Print:
		e.handlePrintCommand(cmd)
	case protocol.Pushing:
		e.publishCurrentState()
	}
}

func (e *Emulator) handlePrintCommand(cmd incomingCommand) {
	switch cmd.Command {
	}
}

func (e *Emulator) publishCurrentState() {
	msg := NewMessageBuilder().
		SetCapability(e.capability).
		SetGcodeState(e.gcodeState).
		Build()

	serialized, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("failed to marshal fabricated message struct: %v", err)
	}

	e.publish(fmt.Sprintf("device/%s/report", e.serial), serialized)
}

func (e *Emulator) PushUpdate() {
	e.publishCurrentState()
}

func (e *Emulator) publish(topic string, payload []byte) {
	_ = e.broker.Publish(topic, payload, false, 0)
}

func (e *Emulator) Stop() {
	e.cancel()
	<-e.done
}
