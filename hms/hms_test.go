package hms_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/torbenconto/bambulabs_api/hms"
	"testing"
)

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name     string
		error    hms.Error
		expected hms.Severity
	}{
		{"Unknown severity", hms.Error{Code: 0}, hms.Unknown},
		{"Fatal severity", hms.Error{Code: 0x10000}, hms.Fatal},
		{"Serious severity", hms.Error{Code: 0x20000}, hms.Serious},
		{"Common severity", hms.Error{Code: 0x30000}, hms.Common},
		{"Info severity", hms.Error{Code: 0x40000}, hms.Info},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.error.GetSeverity())
		})
	}
}

func TestGetModule(t *testing.T) {
	tests := []struct {
		name     string
		error    hms.Error
		expected hms.Module
	}{
		{"Default module", hms.Error{Attribute: 0}, hms.Default},
		{"Mainboard module", hms.Error{Attribute: 0x05000000}, hms.Mainboard},
		{"XCam module", hms.Error{Attribute: 0x0C000000}, hms.XCam},
		{"AMS module", hms.Error{Attribute: 0x07000000}, hms.AMS},
		{"Toolhead module", hms.Error{Attribute: 0x08000000}, hms.Toolhead},
		{"MC module", hms.Error{Attribute: 0x03000000}, hms.MC},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.error.GetModule())
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		error    hms.Error
		expected string
	}{
		{"Valid code", hms.Error{Attribute: 0x12345678, Code: 0x9ABCDEF0}, "1234_5678_9ABC_DEF0"},
		{"Zero attribute and code", hms.Error{Attribute: 0, Code: 0}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.error.GetCode())
		})
	}
}

func TestGetWikiLink(t *testing.T) {
	tests := []struct {
		name     string
		error    hms.Error
		expected string
	}{
		{"Valid link", hms.Error{Attribute: 0x12345678, Code: 0x9ABCDEF0}, "https://wiki.bambulab.com/en/x1/troubleshooting/hmscode/1234_5678_9ABC_DEF0"},
		{"Invalid link (zero attribute/code)", hms.Error{Attribute: 0, Code: 0}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.error.GetWikiLink())
		})
	}
}
