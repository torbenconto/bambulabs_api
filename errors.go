package bambulabs_api

import "errors"

var (
	ErrPrinterExists   = errors.New("printer already present in client")
	ErrPrinterNotFound = errors.New("printer not found")

	ErrLightNotSupported = errors.New("light not supported by this printer model")
	ErrFanNotSupported   = errors.New("fan not supported by this printer model")
)
