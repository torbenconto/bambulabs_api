package bambulabs_api

import (
	"slices"
	"sync"

	"github.com/torbenconto/bambulabs_api/internal/hms"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

// HMSError describes a printer health error, including its code and message.
type HMSError = hms.Error

// HMSSystem holds the latest reported printer health errors.
type HMSSystem struct {
	mu     sync.RWMutex
	errors []hms.Error
}

func NewHMSSystem() *HMSSystem {
	return &HMSSystem{
		errors: make([]hms.Error, 0),
	}
}

// Errors returns an independent snapshot of the current errors.
func (h *HMSSystem) Errors() []HMSError {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return slices.Clone(h.errors)
}

func (h *HMSSystem) apply(errors []hms.Error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.errors = slices.Clone(errors)
}

type HMSDecoder struct{}

func NewHMSDecoder() *HMSDecoder {
	return &HMSDecoder{}
}

func (h *HMSDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil || report.Print.HMSErrors == nil {
		return
	}

	parsedErrors := make([]hms.Error, 0, len(report.Print.HMSErrors))
	for _, err := range report.Print.HMSErrors {
		parsedError := hms.NewError(err.Code, err.Attr)

		parsedErrors = append(parsedErrors, *parsedError)
	}

	p.HMS().apply(parsedErrors)
}
