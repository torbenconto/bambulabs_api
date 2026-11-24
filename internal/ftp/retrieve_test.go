package ftp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"

	goftp "github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/require"
)

type failingDownloadWriter struct{ err error }

func (w failingDownloadWriter) Write([]byte) (int, error) { return 0, w.err }

// downloadServer uses the real FTP control/data protocol on loopback. The
// final reply can fail even after every byte has arrived on the data socket.
func downloadServer(t *testing.T, completion int) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() {
		done <- func() error {
			conn, err := listener.Accept()
			if err != nil {
				return err
			}
			defer conn.Close()
			if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return err
			}
			control := textproto.NewConn(conn)
			if err := control.PrintfLine("220 local test server"); err != nil {
				return err
			}
			var passive net.Listener
			defer func() {
				if passive != nil {
					_ = passive.Close()
				}
			}()
			for {
				line, err := control.ReadLine()
				if err != nil {
					return err
				}
				switch strings.Fields(line)[0] {
				case "EPSV":
					passive, err = net.Listen("tcp", "127.0.0.1:0")
					if err != nil {
						return err
					}
					port := passive.Addr().(*net.TCPAddr).Port
					err = control.PrintfLine("229 Entering Extended Passive Mode (|||%d|)", port)
				case "RETR":
					if passive == nil {
						return errors.New("RETR arrived before EPSV")
					}
					if err = control.PrintfLine("150 Opening data connection"); err != nil {
						return err
					}
					data, acceptErr := passive.Accept()
					if acceptErr != nil {
						return acceptErr
					}
					_, writeErr := io.WriteString(data, "downloaded content")
					_ = data.Close()
					if writeErr != nil {
						return writeErr
					}
					err = control.PrintfLine("%d transfer completion", completion)
				case "QUIT":
					return control.PrintfLine("221 Goodbye")
				default:
					err = control.PrintfLine("200 OK")
				}
				if err != nil {
					return err
				}
			}
		}()
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(6 * time.Second):
			t.Error("local FTP server did not exit")
		}
	})
	return listener.Addr().String()
}

func TestRetrieveReportsTransferCompletionFailure(t *testing.T) {
	writerFailure := errors.New("destination write failed")
	for _, completion := range []int{226, 451} {
		for _, failWriter := range []bool{false, true} {
			t.Run(fmt.Sprintf("completion=%d/writer-failure=%t", completion, failWriter), func(t *testing.T) {
				address := downloadServer(t, completion)
				conn, err := goftp.Dial(address, goftp.DialWithTimeout(5*time.Second))
				require.NoError(t, err)
				client := NewFtpClient(&FtpClientConfig{})
				client.conn = conn
				t.Cleanup(func() { require.NoError(t, client.Close()) })
				var body strings.Builder
				var writer io.Writer = &body
				if failWriter {
					writer = failingDownloadWriter{writerFailure}
				}
				err = client.Retrieve("/part.3mf", writer)
				if failWriter {
					require.ErrorIs(t, err, writerFailure)
				} else {
					require.Equal(t, "downloaded content", body.String())
				}
				if completion != 226 {
					var reply *textproto.Error
					require.ErrorAs(t, err, &reply, "an unsuccessful final FTP reply must fail the download")
					require.Equal(t, completion, reply.Code)
				} else if !failWriter {
					require.NoError(t, err)
				}
			})
		}
	}
}
