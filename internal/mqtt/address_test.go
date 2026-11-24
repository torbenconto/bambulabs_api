package mqtt

import (
	"net"
	"testing"
)

func TestNewMqttClientBrokerAddress(t *testing.T) {
	for _, tc := range []struct {
		host string
		port int
		want string
	}{
		{"127.0.0.1", 8883, "mqtts://127.0.0.1:8883"},
		{"::1", 8883, "mqtts://[::1]:8883"},
		{"2001:db8::20", 18883, "mqtts://[2001:db8::20]:18883"},
	} {
		t.Run(tc.host, func(t *testing.T) {
			client, err := NewMqttClient(&MqttConfig{Host: net.ParseIP(tc.host), Port: tc.port})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			reader := client.client.OptionsReader()
			servers := reader.Servers()
			if len(servers) != 1 {
				t.Fatalf("brokers = %v, want one broker", servers)
			}
			if got := servers[0].String(); got != tc.want {
				t.Errorf("broker = %q, want %q", got, tc.want)
			}
			if _, _, err := net.SplitHostPort(servers[0].Host); err != nil {
				t.Errorf("broker cannot be dialed: %v", err)
			}
		})
	}
}
