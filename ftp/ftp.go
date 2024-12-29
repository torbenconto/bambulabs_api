package ftp

import (
	"bytes"
	"fmt"
	"github.com/secsy/goftp"
	"net"
	"os"
)

type ClientConfig struct {
	Host       net.IP
	Port       int
	Username   string
	AccessCode string
}

type Client struct {
	config *ClientConfig
	conn   *goftp.Client
}

func NewClient(config *ClientConfig) *Client {
	return &Client{
		config: config,
		conn:   nil,
	}
}

func (c *Client) Connect() error {
	config := goftp.Config{
		User:     c.config.Username,
		Password: c.config.AccessCode,
		//TLSConfig:          &tls.Config{InsecureSkipVerify: true},
		TLSMode:            goftp.TLSImplicit,
		DisableEPSV:        false,
		ConnectionsPerHost: 1,
	}

	address := fmt.Sprintf("%s:%d", c.config.Host.String(), c.config.Port)
	conn, err := goftp.DialConfig(config, address)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *Client) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return err
		}
		c.conn = nil
	}
	return nil
}

func (c *Client) StoreFile(path string, data bytes.Buffer) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.Store(path, &data)
}

func (c *Client) RetrieveFile(path string) (bytes.Buffer, error) {
	if c.conn == nil {
		return bytes.Buffer{}, fmt.Errorf("not connected")
	}

	var data bytes.Buffer
	err := c.conn.Retrieve(path, &data)
	return data, err
}

func (c *Client) ListFiles(path string) ([]os.FileInfo, error) {
	if c.conn == nil {
		return []os.FileInfo{}, fmt.Errorf("not connected")
	}

	return c.conn.ReadDir(path)
}

func (c *Client) DeleteFile(path string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.Delete(path)
}
