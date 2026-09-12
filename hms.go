package bambulabs_api

import (
	"sync"

	"github.com/torbenconto/bambulabs_api/internal/hms"
	"github.com/torbenconto/bambulabs_api/internal/protocol"
)

type HMSSystem struct {
	mu     sync.Mutex
	errors []hms.Error
}

func NewHMSSystem() *HMSSystem {
	return &HMSSystem{
		errors: make([]hms.Error, 0),
	}
}

func (h *HMSSystem) apply(errors []hms.Error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.errors = errors
}

type HMSDecoder struct{}

func NewHMSDecoder() *HMSDecoder {
	return &HMSDecoder{}
}

func (h *HMSDecoder) Apply(p *printer, report *protocol.Report) {
	if report.Print == nil {
		return
	}

	var parsedErrors []hms.Error
	for _, err := range report.Print.HMSErrors {
		parsedError := hms.NewError(uint32(err.Code), uint32(err.Attr))

		parsedErrors = append(parsedErrors, *parsedError)
	}

	p.HMS().apply(parsedErrors)

}
