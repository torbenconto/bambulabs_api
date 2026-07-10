package ftp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"sync"

	goftp "github.com/jlaffaye/ftp"
)

type FtpClientConfig struct {
	Host       string
	Port       int
	Username   string
	AccessCode string
}

type FtpClient struct {
	config *FtpClientConfig
	conn   *goftp.ServerConn

	mu        sync.Mutex
	closeOnce sync.Once
	stop      chan struct{}
}

func NewFtpClient(cfg *FtpClientConfig) *FtpClient {
	return &FtpClient{
		config: cfg,
		stop:   make(chan struct{}),
	}
}

func (c *FtpClient) Connect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	conn, err := goftp.Dial(
		addr,
		goftp.DialWithContext(ctx),
		goftp.DialWithTLS(&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.config.Host,
		}),
	)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	if err := c.run(func() error {
		return c.conn.Login(c.config.Username, c.config.AccessCode)
	}); err != nil {
		_ = conn.Quit()
		return fmt.Errorf("ftp login failed: %w", err)
	}

	return nil
}

func (c *FtpClient) List(path string) ([]*goftp.Entry, error) {
	var entries []*goftp.Entry

	if err := c.run(func() error {
		var err error
		entries, err = c.conn.List(path)
		return err
	}); err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *FtpClient) Retrieve(path string, w io.Writer) error {
	return c.run(func() error {
		resp, err := c.conn.Retr(path)
		if err != nil {
			return err
		}
		defer resp.Close()

		_, err = io.Copy(w, resp)
		return err
	})
}

func (c *FtpClient) Store(path string, r io.Reader) error {
	return c.run(func() error {
		return c.conn.Stor(path, r)
	})
}

func (c *FtpClient) Delete(path string) error {
	return c.run(func() error {
		return c.conn.Delete(path)
	})
}

func (c *FtpClient) Done() <-chan struct{} {
	return c.stop
}

func (c *FtpClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	c.closeOnce.Do(func() {
		close(c.stop)
		if c.conn != nil {
			err = c.conn.Quit()
			c.conn = nil
		}
	})

	return err
}

func (c *FtpClient) run(fn func() error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.stop:
		return ErrClosed
	default:
	}

	if c.conn == nil {
		return ErrClosed
	}

	return fn()
}
