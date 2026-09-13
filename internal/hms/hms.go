package hms

import (
	"fmt"
	"strings"
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
		return fmt.Sprintf("HMS_%04X_%04X_%04X_%04X", attrHigh, attrLow, codeHigh, codeLow)
	}
	return ""
}

func (e Error) Error() string {
	// The generated wiki table uses hyphens between the numeric groups.
	key := "HMS_" + strings.ReplaceAll(strings.TrimPrefix(e.GetCode(), "HMS_"), "_", "-")
	if msg, ok := HmsErrors[key]; ok {
		return msg
	}

	return e.GetCode()
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
