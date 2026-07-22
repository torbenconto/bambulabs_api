package bambulabs_api

import "errors"

var (
	ErrPrinterExists   = errors.New("printer already present in client")
	ErrPrinterNotFound = errors.New("printer not found")

	ErrLightNotAvalible = errors.New("this light is not currently avalible, it may not be supported by your printer")

	ErrFTPUnavailable = errors.New("ftp connection unavailable")
)
