package hms

import (
	"fmt"
	"strconv"
)

type Severity string

var Severities = map[int]Severity{
	0: Unknown,
	1: Fatal,
	2: Serious,
	3: Common,
	4: Info,
}

const (
	Unknown Severity = "unknown"
	Fatal   Severity = "fatal"
	Serious Severity = "serious"
	Common  Severity = "common"
	Info    Severity = "info"
)

type Module int

var Modules = map[Module]string{
	Default:   "Default",
	Mainboard: "Mainboard",
	XCam:      "XCam",
	AMS:       "AMS",
	Toolhead:  "Toolhead",
	MC:        "MC",
}

const (
	Default   Module = 0x00
	Mainboard Module = 0x05
	XCam      Module = 0x0C
	AMS       Module = 0x07
	Toolhead  Module = 0x08
	MC        Module = 0x03
)

type Error struct {
	Attribute int `json:"attr"`
	Code      int `json:"code"`
}

func (e Error) GetSeverity() Severity {
	uint_code := e.Code >> 16
	if e.Code > 0 {
		if _, ok := Severities[uint_code]; ok {
			return Severities[uint_code]
		}
	}
	return Unknown
}

func (e Error) GetModule() Module {
	uint_attr := (e.Attribute >> 24) & 0xff
	if e.Attribute > 0 {
		if _, ok := Modules[Module(uint_attr)]; ok {
			return Module(uint_attr)
		}
	}
	return Default
}

func (e Error) GetCode() string {
	if e.Attribute > 0 && e.Code > 0 {
		attrHigh := (e.Attribute >> 16) & 0xffff
		attrLow := e.Attribute & 0xffff
		codeHigh := (e.Code >> 16) & 0xffff
		codeLow := e.Code & 0xffff
		return fmt.Sprintf("%04X_%04X_%04X_%04X", attrHigh, attrLow, codeHigh, codeLow)
	}
	return ""
}

func (e Error) GetGenericErrorCode() string {
	code := e.GetCode()
	if len(code) != 19 {
		return ""
	}

	code1, _ := strconv.ParseInt(code[0:4], 16, 32)
	code2, _ := strconv.ParseInt(code[5:9], 16, 32)
	code3, _ := strconv.ParseInt(code[10:14], 16, 32)
	code4, _ := strconv.ParseInt(code[15:19], 16, 32)

	amsCode := fmt.Sprintf("%04X_%04X_%04X_%04X",
		code1&0xfff8,
		code2&0xf8ff,
		code3,
		code4,
	)

	if _, ok := AMSErrors[amsCode]; ok {
		return amsCode
	}

	return fmt.Sprintf("%04X_%04X_%04X_%04X",
		code1,
		code2,
		code3,
		code4,
	)
}

func (e Error) GetWikiLink() string {
	if e.Attribute > 0 && e.Code > 0 {
		return fmt.Sprintf("https://wiki.bambulab.com/en/x1/troubleshooting/hmscode/%s", e.GetGenericErrorCode())
	}
	return ""
}
