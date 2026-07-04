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

	mu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	closeOnce sync.Once
}

func NewFtpClient(parent context.Context, cfg *FtpClientConfig) (*FtpClient, error) {
	lifecycleCtx, cancel := context.WithCancel(parent)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := goftp.Dial(addr, goftp.DialWithContext(parent), goftp.DialWithTLS(&tls.Config{
		InsecureSkipVerify: true,
		ServerName:         cfg.Host,
		ClientSessionCache: tls.NewLRUClientSessionCache(0),
	}))
	if err != nil {
		cancel()
		return nil, err
	}

	c := &FtpClient{
		config: cfg,
		conn:   conn,
		ctx:    lifecycleCtx,
		cancel: cancel,
	}

	if err := c.Login(parent); err != nil {
		cancel()
		_ = conn.Quit()
		return nil, fmt.Errorf("ftp login failed: %w", err)
	}

	return c, nil
}

func (c *FtpClient) Login(ctx context.Context) error {
	return c.run(ctx, func() error {
		return c.conn.Login(c.config.Username, c.config.AccessCode)
	})
}

func (c *FtpClient) List(ctx context.Context, path string) ([]*goftp.Entry, error) {
	var entries []*goftp.Entry
	err := c.run(ctx, func() error {
		var innerErr error
		entries, innerErr = c.conn.List(path)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *FtpClient) Retrieve(ctx context.Context, path string, w io.Writer) error {
	return c.run(ctx, func() error {
		resp, err := c.conn.Retr(path)
		if err != nil {
			return err
		}
		defer resp.Close()

		_, err = io.Copy(w, resp)
		return err
	})
}

func (c *FtpClient) Store(ctx context.Context, path string, r io.Reader) error {
	return c.run(ctx, func() error {
		return c.conn.Stor(path, r)
	})
}

func (c *FtpClient) Delete(ctx context.Context, path string) error {
	return c.run(ctx, func() error {
		return c.conn.Delete(path)
	})
}

func (c *FtpClient) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *FtpClient) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.cancel()
		err = c.conn.Quit()
	})
	return err
}

func (c *FtpClient) run(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}
