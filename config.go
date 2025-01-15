package bambulabs_api

import "net"

type PrinterConfig struct {
	IP           net.IP
	AccessCode   string
	SerialNumber string
}

func NewPrinterConfig(IP net.IP, AccessCode, SerialNumber string) *PrinterConfig {
	return &PrinterConfig{
		IP,
		AccessCode,
		SerialNumber,
	}
}
