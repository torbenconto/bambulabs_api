package mqtt

import (
	"context"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type readyToken struct {
	paho.Token
	done chan struct{}
}

func (t *readyToken) Done() <-chan struct{} { return t.done }
func (t *readyToken) Error() error          { return nil }

type recordingTransport struct {
	paho.Client
	token     paho.Token
	connects  int
	publishes int
}

func (c *recordingTransport) Connect() paho.Token {
	c.connects++
	return c.token
}

func (c *recordingTransport) Publish(string, byte, bool, interface{}) paho.Token {
	c.publishes++
	return c.token
}

func TestStoppedOperationsNeverReachTransport(t *testing.T) {
	for _, condition := range []string{"canceled", "expired", "closed"} {
		t.Run(condition, func(t *testing.T) {
			ctx := t.Context()
			stop := make(chan struct{})
			var want error
			switch condition {
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = context.Canceled
			case "expired":
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
				want = context.DeadlineExceeded
			case "closed":
				close(stop)
				want = ErrClosed
			}
			done := make(chan struct{})
			close(done)
			transport := &recordingTransport{token: &readyToken{done: done}}
			client := &MqttClient{config: &MqttConfig{SerialNumber: "LOCAL"}, client: transport, stop: stop}
			connectErr := client.Connect(ctx)
			publishErr := client.Publish(ctx, protocol.NewCommand(protocol.Pushing).WithCommand("pushall"))
			require.Zero(t, transport.connects, "a canceled or closed client must not start a connection")
			require.Zero(t, transport.publishes, "a canceled or closed request must not send printer commands")
			require.ErrorIs(t, connectErr, want)
			require.ErrorIs(t, publishErr, want)
		})
	}
}

func TestActiveOperationsStillReachTransport(t *testing.T) {
	done := make(chan struct{})
	close(done)
	transport := &recordingTransport{token: &readyToken{done: done}}
	client := &MqttClient{config: &MqttConfig{SerialNumber: "LOCAL"}, client: transport, stop: make(chan struct{})}
	require.NoError(t, client.Connect(t.Context()))
	require.NoError(t, client.Publish(t.Context(), protocol.NewCommand(protocol.Pushing).WithCommand("pushall")))
	require.Equal(t, 1, transport.connects)
	require.Equal(t, 1, transport.publishes)
}
