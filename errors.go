package bambulabs_api

import "errors"

var (
	ErrPrinterExists   = errors.New("printer already present in client")
	ErrPrinterNotFound = errors.New("printer not found")
)
