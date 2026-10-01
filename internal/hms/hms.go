package hms

import (
	"fmt"
)

type Error struct {
	Attribute uint32 `json:"attribute"`
	Code      uint32 `json:"code"`
}

func NewError(code, attribute uint32) *Error {
	return &Error{
		Attribute: attribute,
		Code:      code,
	}
}

func (e Error) GetCode() string {
	if e.Attribute > 0 && e.Code > 0 {
		attrHigh := (e.Attribute >> 16) & 0xffff
		attrLow := e.Attribute & 0xffff
		codeHigh := (e.Code >> 16) & 0xffff
		codeLow := e.Code & 0xffff
		// The generated wiki table uses hyphens between the numeric groups.
		return fmt.Sprintf("HMS_%04X_%04X_%04X_%04X", attrHigh, attrLow, codeHigh, codeLow)
	}
	return ""
}

func (e Error) Error() string {
	eCode := e.GetCode()
	if msg, ok := HmsErrors[eCode]; ok {
		// NOTE: just picking the first info string in the list, though there may be more than one
		return msg[0]
	}
	return eCode
}

type Module uint8

const (
	ModuleDefault   Module = 0x00
	ModuleMainboard Module = 0x05
	ModuleXCam      Module = 0x0C
	ModuleAMS       Module = 0x07
	ModuleToolhead  Module = 0x08
	ModuleMC        Module = 0x03
)

type Severity uint8

const (
	SeverityInvalid      Severity = iota // 0000
	SeverityError                        // 0001
	SeverityWarning                      // 0002
	SeverityNotification                 // 0003
)
