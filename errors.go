package bambulabs_api

import "errors"

var (
	ErrPrinterExists   = errors.New("printer already present in client")
	ErrPrinterNotFound = errors.New("printer not found")

	ErrLightUnavailable = errors.New("this light is not currently avalible, it may not be supported by your printer")
	ErrFanUnavailable   = errors.New("this fan is not currently avalible, it may not be supported by your printer")

	ErrInvalidFanPercent = errors.New("fan percent must be between 0 and 100")
	ErrInvalidLightMode  = errors.New("invalid light mode")

	ErrFTPUnavailable = errors.New("ftp connection unavailable")
)
