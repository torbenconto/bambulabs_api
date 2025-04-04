package bambulabs_api

import (
	"fmt"
	"sync"
)

type PrinterPool struct {
	printers sync.Map

	// List of serial numbers in the order they were added, used for sequential operations where order matters
	order []string
}

func NewPrinterPool() *PrinterPool {
	return &PrinterPool{}
}

func (p *PrinterPool) ConnectAll() error {
	return p.ExecuteAll(func(printer *Printer) error {
		return printer.Connect()
	})
}

func (p *PrinterPool) ConnectAllCamera() error {
	return p.ExecuteAll(func(printer *Printer) error {
		return printer.ConnectCamera()
	})
}

func (p *PrinterPool) DisconnectAll() error {
	return p.ExecuteAll(func(printer *Printer) error {
		return printer.Disconnect()
	})
}

func (p *PrinterPool) DisconnectAllCamera() error {
	return p.ExecuteAll(func(printer *Printer) error {
		return printer.DisconnectCamera()
	})
}

func (p *PrinterPool) AddPrinter(config *PrinterConfig) {
	printer := NewPrinter(config)

	p.printers.Store(config.SerialNumber, printer)
	p.order = append(p.order, config.SerialNumber)
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

	// Remove the serial number from the order slice
	for i, serial := range p.order {
		if serial == serialNumber {
			p.order = append(p.order[:i], p.order[i+1:]...)
			break
		}
	}
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

// ExecuteOnN performs a printer operation on a subset of printers in the pool concurrently.
func (p *PrinterPool) ExecuteOnN(operation func(*Printer) error, n []int) error {
	if len(n) == 0 {
		return fmt.Errorf("no printer selected")
	}

	allPrinters := p.GetPrinters()

	var wg sync.WaitGroup
	errChan := make(chan error, len(n))

	for _, i := range n {
		if i < 0 || i >= len(allPrinters) {
			return fmt.Errorf("index %d out of range", i)
		}
		printer := allPrinters[i]

		wg.Add(1)
		go func(p *Printer) {
			defer wg.Done()
			if err := operation(p); err != nil {
				errChan <- fmt.Errorf("operation on printer %s error: %w", p.serial, err)
			}
		}(printer)
	}

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

// ExecuteOnNSequentially performs a printer operation on a subset of printers in the pool sequentially.
func (p *PrinterPool) ExecuteOnNSequentially(operation func(*Printer) error, n []int) error {
	if len(n) == 0 {
		return fmt.Errorf("no printer selected")
	}

	allPrinters := p.GetPrinters()

	var result error

	for _, i := range n {
		if i < 0 || i >= len(allPrinters) {
			return fmt.Errorf("index %d out of range", i)
		}
		printer := allPrinters[i]

		if err := operation(printer); err != nil {
			result = fmt.Errorf("operation on printer %s error: %w", printer.serial, err)
			break
		}
	}

	return result
}

// ExecuteAllSequentially performs a printer operation on all printers in the pool sequentially.
// This is useful for operations that need to be performed in a specific order such as a light show or any operation based on physical constraints.
func (p *PrinterPool) ExecuteAllSequentially(operation func(*Printer) error) error {
	var result error

	for _, serial := range p.order {
		printerInterface, ok := p.printers.Load(serial)
		if !ok {
			continue
		}
		printer, ok := printerInterface.(*Printer)
		if !ok {
			continue
		}

		if err := operation(printer); err != nil {
			result = fmt.Errorf("operation on printer %s error: %w", printer.serial, err)
			break
		}
	}

	return result
}

// DataAll collects data from all printers in the pool and returns it as a map where the serial number of a printer is the key and maps to a reference to its data.
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
	printerInterface, ok := p.printers.Load(serial)
	if !ok {
		return nil, fmt.Errorf("printer with serial %s not found", serial)
	}

	printer, ok := printerInterface.(*Printer)
	if !ok {
		return nil, fmt.Errorf("invalid printer type for serial %s", serial)
	}

	return printer, nil
}
