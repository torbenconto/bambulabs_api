package bambulabs_api

import (
	"context"
	"fmt"
	"sync"
)

// Model represents the printers model number
type Model uint

// Model names and capabilites sourced from https://bambulab.com/en/compare
const (
	ModelUnknown Model = iota
	ModelA1Mini
	ModelA1
	ModelA2L
	ModelP1S
	ModelP2S
	ModelX1E
	ModelX1C
	ModelH2S
	ModelH2D
	ModelH2DPro
	ModelH2
	ModelH2C
	ModelX2D
)

// Core client struct, v0.1.6 and below were sloppy and required manual control of printer structs and the pool abstraction was just layered on top
// This aims to fix those issues by providing a unified interface for printer interaction
type Client struct {
	printers sync.Map

	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(parent context.Context) *Client {
	ctx, cancel := context.WithCancel(parent)
	return &Client{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) Add(cfg Config) (Printer, error) {
	if _, ok := c.printers.Load(cfg.SerialNumber); ok {
		return nil, ErrPrinterExists
	}

	p, err := NewPrinter(c.ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", cfg.SerialNumber, err)
	}

	if _, exists := c.printers.LoadOrStore(cfg.SerialNumber, p); exists {
		_ = p.Close()
		return nil, ErrPrinterExists
	}

	return p, nil
}

func (c *Client) Load(serial string) (Printer, error) {
	if p, ok := c.printers.Load(serial); ok {
		return p.(Printer), nil
	}
	return nil, ErrPrinterNotFound
}

func (c *Client) Remove(serial string) error {
	v, ok := c.printers.LoadAndDelete(serial)
	if !ok {
		return ErrPrinterNotFound
	}
	p := v.(Printer)
	return p.Close()
}

func (c *Client) Range(fn func(Printer) bool) {
	c.printers.Range(func(_, value any) bool {
		return fn(value.(Printer))
	})
}

func (c *Client) Close() error {
	c.cancel()

	var firstErr error
	c.printers.Range(func(key, value any) bool {
		p := value.(Printer)
		if err := p.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		c.printers.Delete(key)
		return true
	})

	return firstErr
}
