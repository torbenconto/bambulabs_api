package ftp

import (
	"crypto/tls"
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

// Connect connects to the ftp server of a bambu printer.
// This function is working and has been tested on:
// - [x] X1 Carbon
func (c *Client) Connect() error {
	config := goftp.Config{
		User:     c.config.Username,
		Password: c.config.AccessCode,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			ClientSessionCache: tls.NewLRUClientSessionCache(0),
			ServerName:         c.config.Host.String(),
		},

		TLSMode: goftp.TLSImplicit,
	}

	address := fmt.Sprintf("%s:%d", c.config.Host.String(), c.config.Port)
	conn, err := goftp.DialConfig(config, address)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// Disconnect disconnects from the ftp server.
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

// StoreFile stores file "file" into "path" on the server.
func (c *Client) StoreFile(path string, file os.File) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.Store(path, &file)
}

// RetrieveFile retrieves file "path" on the server and returns it's contents as a buffer.
func (c *Client) RetrieveFile(path string) (os.File, error) {
	if c.conn == nil {
		return os.File{}, fmt.Errorf("not connected")
	}

	var data os.File
	err := c.conn.Retrieve(path, &data)
	return data, err
}

// ListDir lists a given directory on the server and returns a list of the files and directories it contains as an array of os.FileInfo.
func (c *Client) ListDir(path string) ([]os.FileInfo, error) {
	if c.conn == nil {
		return []os.FileInfo{}, fmt.Errorf("not connected")
	}

	return c.conn.ReadDir(path)
}

// DeleteFile deletes file "path" on the ser
func (c *Client) DeleteFile(path string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.Delete(path)
}
