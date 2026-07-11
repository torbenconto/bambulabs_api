// Copyright (c) 2026 Torben Conto
//
// Package bambulabs_api provides a Go interface for communicating with
// Bambu Lab 3D printers over their local network interfaces.
//
// The library supports printer telemetry and control through MQTT, as well
// as file management through FTP. It is designed for local LAN usage and
// does not require cloud connectivity.
//
// This project is an independent, unofficial library and is not affiliated
// with, endorsed by, sponsored by, or otherwise associated with Bambu Lab.
//
// Users are responsible for ensuring that any operations performed through
// this library are safe for their hardware and comply with applicable
// software and hardware terms.
//
// Package stability and printer compatibility may vary depending on printer
// firmware versions and available LAN features.
package bambulabs_api

import (
	"context"
	"fmt"
	"sync"
)

// Model represents the printers model number
//
// The model is used to determine printer capabilities, including available
// lights, fans, and other hardware features. Use ModelUnknown when the exact
// printer model is not known.
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

// Client manages connections to one or more Bambu Lab printers.
//
// A Client owns the lifetime of all printers added to it. Closing a Client
// closes all associated printer connections and releases resources.
//
// The Client is safe for concurrent use.
//
// v0.1.6 and below were sloppy and required manual control of printer structs and the pool abstraction was just layered on top
// This aims to fix those issues by providing a unified interface for printer interaction
type Client struct {
	mu       sync.Mutex
	printers sync.Map // map[printer serial number]Printer

	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient creates a new printer client using parent as its lifetime context.
//
// When parent is canceled or Close is called, the client shuts down and all
// managed printers are closed.
func NewClient(parent context.Context) *Client {
	ctx, cancel := context.WithCancel(parent)
	return &Client{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Add connects to a [Printer] described by cfg and adds it to the client's
// managed printer collection.
//
// The MQTT connection must be established successfully for Add to succeed.
// FTP connectivity is optional; if FTP setup fails, the printer is still added
// but file operations will return [ErrFTPUnavailable].
//
// Add returns [ErrPrinterExists] if a printer with the same serial number is
// already managed by the client.
func (c *Client) Add(cfg Config) (Printer, error) {
	c.mu.Lock() // lock printers to prevent races
	defer c.mu.Unlock()

	if _, ok := c.printers.Load(cfg.SerialNumber); ok { // check if the printer already exists
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

// Load retrieves a managed printer by its serial number.
//
// Load returns [ErrPrinterNotFound] if no printer with the given serial number
// exists in the client.
func (c *Client) Load(serial string) (Printer, error) {
	if p, ok := c.printers.Load(serial); ok {
		return p.(Printer), nil
	}
	return nil, ErrPrinterNotFound
}

// Remove removes a printer from the client and closes its connections.
//
// Remove returns [ErrPrinterNotFound] if no printer with the given serial number
// exists.
func (c *Client) Remove(serial string) error {
	v, ok := c.printers.LoadAndDelete(serial)
	if !ok {
		return ErrPrinterNotFound
	}
	p := v.(Printer)
	return p.Close()
}

// Range calls fn sequentially for each printer currently managed by the client.
//
// Returning false from fn stops iteration early.
//
// The order of iteration is not guaranteed.
func (c *Client) Range(fn func(Printer) bool) {
	c.printers.Range(func(_, value any) bool {
		return fn(value.(Printer))
	})
}

// Close shuts down the client and closes all managed printers.
//
// Close cancels the client's lifetime context, causing background operations
// to stop, and releases all resources owned by the client.
//
// Calling Close multiple times is safe.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

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
