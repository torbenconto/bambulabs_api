package bambulabs_api

import "log/slog"

type PrinterConfig struct {
	Host         string
	AccessCode   string
	SerialNumber string

	Logger *slog.Logger
}
