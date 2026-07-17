package emulator

import (
	"context"
	"fmt"
	"os"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/torbenconto/bambulabs_api"
)

type Emulator struct {
	cancel context.CancelFunc
	done   chan struct{}

	broker *mochi.Server

	serial string
	report []byte
}

func Start(ctx context.Context, cfg *bambulabs_api.Config, port int, reportFile string) (*Emulator, error) {
	ctx, cancel := context.WithCancel(ctx)

	report, err := os.ReadFile(reportFile)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("read report: %w", err)
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
		report: report,
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
	_ packets.Packet,
) {
	e.publishReport()
}

func (e *Emulator) publishReport() {
	_ = e.broker.Publish(
		fmt.Sprintf("device/%s/report", e.serial),
		e.report,
		false,
		0,
	)
}

func (e *Emulator) PushUpdate() {
	e.publishReport()
}

func (e *Emulator) Stop() {
	e.cancel()
	<-e.done
}
