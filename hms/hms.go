package hms

import (
	"fmt"
	"strings"
)

type Error struct {
	Attribute uint32 `json:"attribute"`
	Code      uint32 `json:"code"`
}

func NewError(code string) *Error {
	code, _ = strings.CutPrefix(code, "HMS_")

	var attrHigh, attrLow uint32
	var codeHigh, codeLow uint32

	_, err := fmt.Sscanf(code, "%04x_%04x_%04x_%04x", &attrHigh, &attrLow, &codeHigh, &codeLow)
	if err != nil {
		return nil
	}

	parsedAttr := (attrHigh << 16) | attrLow
	parsedCode := (codeHigh << 16) | codeLow

	return &Error{
		Attribute: parsedAttr,
		Code:      parsedCode,
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
	if msg, ok := HmsErrors[e.GetCode()]; ok {
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
