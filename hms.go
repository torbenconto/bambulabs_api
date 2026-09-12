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
	return &HMSSystem{}
}

func (h *HMSSystem) apply(errors []hms.Error)

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
		// TODO: actuall parse error
		_ = err.Attr
		parsedErrors = append(parsedErrors)
	}

	p.HMS().apply(parsedErrors)

}
