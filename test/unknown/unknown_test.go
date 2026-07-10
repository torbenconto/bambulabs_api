package x1_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/torbenconto/bambulabs_api"
	"github.com/torbenconto/bambulabs_api/internal/emulator"
)

var (
	cfg = bambulabs_api.Config{
		Host:         net.ParseIP("127.0.0.1"),
		MQTTPort:     mqttPort,
		Model:        bambulabs_api.ModelX1C,
		AccessCode:   "test1234",
		SerialNumber: "BBLX1C0001",
	}
	emu      *emulator.Emulator
	mqttPort = 18883
)

func TestMain(m *testing.M) {
	var err error
	emu, err = emulator.Start(context.Background(), &cfg, mqttPort)
	if err != nil {
		panic("start emulator: " + err.Error())
	}
	code := m.Run()
	emu.Stop()
	os.Exit(code)
}

func client(t *testing.T) (*bambulabs_api.Client, bambulabs_api.Printer) {
	t.Helper()
	c := bambulabs_api.NewClient(context.Background())
	t.Cleanup(func() { c.Close() })
	p, err := c.Add(cfg)
	if err != nil {
		t.Fatalf("add printer: %v", err)
	}
	return c, p
}

func TestConnect(t *testing.T) {
	_, p := client(t)
	if p.Serial() != cfg.SerialNumber {
		t.Errorf("serial: got %q want %q", p.Serial(), cfg.SerialNumber)
	}
}

func TestUnsolicitedUpdate(t *testing.T) {
	_, p := client(t)

	emu.PushUpdate()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for unsolicited update")
		case <-time.After(50 * time.Millisecond):
			if _, ok := p.State(); ok {
				return
			}
		}
	}
}

func TestSolicitedUpdate(t *testing.T) {
	_, p := client(t)

	if err := p.RequestUpdate(context.Background()); err != nil {
		t.Fatalf("request update: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for solicited update")
		case <-time.After(50 * time.Millisecond):
			if _, ok := p.State(); ok {
				return
			}
		}
	}
}
