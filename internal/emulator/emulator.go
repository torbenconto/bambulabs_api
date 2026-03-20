package emulator

import (
	"context"
	"fmt"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/torbenconto/bambulabs_api"
)

type Emulator struct {
	cancel context.CancelFunc
	port   int
	host   string
	broker *mochi.Server
	done   chan struct{}
}

func Start(ctx context.Context, cfg *bambulabs_api.Config, port int) (*Emulator, error) {
	ctx, cancel := context.WithCancel(ctx)

	server := mochi.New(nil)

	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		cancel()
		return nil, fmt.Errorf("add auth hook %v", err)
	} // allow all users r/w to any topic

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "torbenconto/bambulabs_api/emulator",
		Address: fmt.Sprintf("127.0.0.1:%d", port),
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
		host:   "127.0.0.1",
		port:   port,
		broker: server,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	go emu.run(ctx)

	return emu, nil
}

func (e *Emulator) run(ctx context.Context) {
	defer close(e.done)
	<-ctx.Done()
	_ = e.broker.Close()
}

func (e *Emulator) Stop() {
	e.cancel()
	<-e.done
}
