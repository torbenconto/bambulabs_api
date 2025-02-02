package hms

import (
	"fmt"
	"strconv"
)

type HMSSeverity string

var HMSSeverities = map[int]HMSSeverity{
	0: Unknown,
	1: Fatal,
	2: Serious,
	3: Common,
	4: Info,
}

const (
	Unknown HMSSeverity = "unknown"
	Fatal   HMSSeverity = "fatal"
	Serious HMSSeverity = "serious"
	Common  HMSSeverity = "common"
	Info    HMSSeverity = "info"
)

type HMSModule int

var HMSModules = map[HMSModule]string{
	Default:   "Default",
	Mainboard: "Mainboard",
	XCam:      "XCam",
	AMS:       "AMS",
	Toolhead:  "Toolhead",
	MC:        "MC",
}

const (
	Default   HMSModule = 0x00
	Mainboard HMSModule = 0x05
	XCam      HMSModule = 0x0C
	AMS       HMSModule = 0x07
	Toolhead  HMSModule = 0x08
	MC        HMSModule = 0x03
)

type HMSError struct {
	Attribute int `json:"attr"`
	Code      int `json:"code"`
}

func (e HMSError) GetServerity() HMSSeverity {
	uint_code := e.Code >> 16
	if e.Code > 0 {
		if _, ok := HMSSeverities[uint_code]; ok {
			return HMSSeverities[uint_code]
		}
	}
	return Unknown
}

func (e HMSError) GetModule() HMSModule {
	uint_attr := (e.Attribute >> 24) & 0xff
	if e.Attribute > 0 {
		if _, ok := HMSModules[HMSModule(uint_attr)]; ok {
			return HMSModule(uint_attr)
		}
	}
	return Default
}

func (e HMSError) GetHMSCode() string {
	if e.Attribute > 0 && e.Code > 0 {
		attrHigh := (e.Attribute >> 16) & 0xffff
		attrLow := e.Attribute & 0xffff
		codeHigh := (e.Code >> 16) & 0xffff
		codeLow := e.Code & 0xffff
		return fmt.Sprintf("%04X_%04X_%04X_%04X", attrHigh, attrLow, codeHigh, codeLow)
	}
	return ""
}

func (e HMSError) GetGenericHMSErrorCode() string {
	hmsCode := e.GetHMSCode()
	if len(hmsCode) != 19 {
		return ""
	}

	code1, _ := strconv.ParseInt(hmsCode[0:4], 16, 32)
	code2, _ := strconv.ParseInt(hmsCode[5:9], 16, 32)
	code3, _ := strconv.ParseInt(hmsCode[10:14], 16, 32)
	code4, _ := strconv.ParseInt(hmsCode[15:19], 16, 32)

	amsCode := fmt.Sprintf("%04X_%04X_%04X_%04X",
		code1&0xfff8,
		code2&0xf8ff,
		code3,
		code4,
	)

	if _, ok := HMSAMSErrors[amsCode]; ok {
		return amsCode
	}

	return fmt.Sprintf("%04X_%04X_%04X_%04X",
		code1,
		code2,
		code3,
		code4,
	)
}

func (e HMSError) GetWikiLink() string {
	if e.Attribute > 0 && e.Code > 0 {
		return fmt.Sprintf("https://wiki.bambulab.com/en/x1/troubleshooting/hmscode/%s", e.GetGenericHMSErrorCode())
	}
	return ""
}
