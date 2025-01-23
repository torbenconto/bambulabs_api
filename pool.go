package bambulabs_api

import (
	"fmt"
	"sync"
)

type PrinterPool struct {
	mu       sync.Mutex
	printers sync.Map
}

func NewPrinterPool() *PrinterPool {
	return &PrinterPool{}
}

func (p *PrinterPool) ConnectAll() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	p.printers.Range(func(_, value interface{}) bool {
		printer, ok := value.(*Printer)
		if !ok {
			return false
		}

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()
			err := p.Connect()
			if err != nil {
				errChan <- fmt.Errorf("failed to connect to printer %s: %w", p.serial, err)
			}
		}(printer)

		return true
	})

	wg.Wait()
	close(errChan)

	var result error
	for err := range errChan {
		if result == nil {
			result = err
		} else {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}

	return result
}

func (p *PrinterPool) ConnectMqttAll() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	p.printers.Range(func(_, value interface{}) bool {
		printer, ok := value.(*Printer)
		if !ok {
			return false
		}

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()
			err := p.mqttClient.Connect()
			if err != nil {
				errChan <- fmt.Errorf("failed to connect to MQTT client of printer %s: %w", p.serial, err)
			}
		}(printer)

		return true
	})

	wg.Wait()
	close(errChan)

	var result error
	for err := range errChan {
		if result == nil {
			result = err
		} else {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}

	return result
}

func (p *PrinterPool) ConnectFtpAll() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	p.printers.Range(func(_, value interface{}) bool {
		printer, ok := value.(*Printer)
		if !ok {
			return false
		}

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()
			err := p.ftpClient.Connect()
			if err != nil {
				errChan <- fmt.Errorf("failed to connect to FTP client of printer %s: %w", p.serial, err)
			}
		}(printer)

		return true
	})

	wg.Wait()
	close(errChan)

	var result error
	for err := range errChan {
		if result == nil {
			result = err
		} else {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}

	return result
}

func (p *PrinterPool) DisconnectAll() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	p.printers.Range(func(_, value interface{}) bool {
		printer, ok := value.(*Printer)
		if !ok {
			return false
		}

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()
			if err := p.Disconnect(); err != nil {
				errChan <- fmt.Errorf("printer %s disconnect error: %w", p.serial, err)
			}
		}(printer)
		return true
	})

	wg.Wait()
	close(errChan)

	var result error
	for err := range errChan {
		if result == nil {
			result = err
		} else {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}

	return result
}

func (p *PrinterPool) AddPrinter(config *PrinterConfig) {
	printer := NewPrinter(config)

	p.printers.Store(config.SerialNumber, printer)
}

func (p *PrinterPool) GetPrinter(serialNumber string) *Printer {
	printer, _ := p.printers.Load(serialNumber)

	return printer.(*Printer)
}

func (p *PrinterPool) GetPrinters() []*Printer {
	var printers []*Printer

	p.printers.Range(func(_, value interface{}) bool {
		printers = append(printers, value.(*Printer))
		return true
	})

	return printers
}

func (p *PrinterPool) RemovePrinter(serialNumber string) {
	p.printers.Delete(serialNumber)
}

// ExecuteAll performs a printer operation on all printers in the pool concurrently.
func (p *PrinterPool) ExecuteAll(operation func(*Printer) error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	p.printers.Range(func(_, value interface{}) bool {
		printer, ok := value.(*Printer)
		if !ok {
			return false
		}

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()
			if err := operation(p); err != nil {
				errChan <- fmt.Errorf("operation on printer %s error: %w", p.serial, err)
			}
		}(printer)
		return true
	})

	wg.Wait()
	close(errChan)

	var result error
	for err := range errChan {
		if result == nil {
			result = err
		} else {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}

	return result
}

// DataAll collects data from all printers in the pool and returns it as a map where the serial number of a printer is the key and maps to a reference to it's data.
func (p *PrinterPool) DataAll() (map[string]*Data, error) {
	var wg sync.WaitGroup
	result := sync.Map{}
	errCh := make(chan error, 100)

	p.printers.Range(func(_, value interface{}) bool {
		printer, ok := value.(*Printer)
		if !ok {
			return false
		}

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()

			data, err := p.Data()
			if err != nil {
				errCh <- fmt.Errorf("printer %s data error: %w", p.serial, err)
				return
			}

			result.Store(p.serial, &data)
		}(printer)
		return true
	})

	wg.Wait()
	close(errCh)

	var resultErr error
	for err := range errCh {
		if resultErr == nil {
			resultErr = err
		} else {
			resultErr = fmt.Errorf("%v; %w", resultErr, err)
		}
	}

	// Convert sync.Map to a regular map
	finalResult := make(map[string]*Data)
	result.Range(func(key, value interface{}) bool {
		serial, _ := key.(string)
		data, _ := value.(*Data)
		finalResult[serial] = data
		return true
	})

	return finalResult, resultErr
}

// At retrieves a printer by its serial number.
func (p *PrinterPool) At(serial string) (*Printer, error) {
	var printer *Printer
	var found bool

	p.printers.Range(func(key, value interface{}) bool {
		if key == serial {
			printer, found = value.(*Printer)
			return false
		}
		return true
	})

	if !found {
		return nil, fmt.Errorf("printer with serial %s not found", serial)
	}

	return printer, nil
}
