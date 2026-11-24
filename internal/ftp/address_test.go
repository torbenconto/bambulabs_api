package ftp

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestConnectDialsConfiguredAddress(t *testing.T) {
	for _, tc := range []struct {
		name    string
		network string
		address string
	}{
		{"IPv4", "tcp4", "127.0.0.1:0"},
		{"IPv6", "tcp6", "[::1]:0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen(tc.network, tc.address)
			if err != nil {
				if tc.network == "tcp6" {
					t.Skipf("IPv6 loopback is unavailable: %v", err)
				}
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = listener.Close() })
			accepted := make(chan error, 1)
			go func() {
				conn, err := listener.Accept()
				if err == nil {
					// Reaching this socket proves that Connect formed a valid
					// address. No TLS handshake or FTP login is needed here.
					_ = conn.Close()
				}
				accepted <- err
			}()

			addr := listener.Addr().(*net.TCPAddr)
			client := NewFtpClient(&FtpClientConfig{Host: addr.IP.String(), Port: addr.Port})
			t.Cleanup(func() { _ = client.Close() })
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			connectErr := client.Connect(ctx)
			if connectErr == nil {
				t.Fatal("Connect succeeded after the test listener closed the connection")
			}
			select {
			case err := <-accepted:
				if err != nil {
					t.Fatalf("accept configured endpoint: %v (Connect: %v)", err, connectErr)
				}
			case <-ctx.Done():
				t.Fatalf("Connect did not reach %s: %v", addr, connectErr)
			}
		})
	}
}
