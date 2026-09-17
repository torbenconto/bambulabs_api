package hms

import "testing"

func TestNewErrorUsesCodeAndAttributeDirectly(t *testing.T) {
	const (
		code      uint32 = 0x0300800A
		attribute uint32 = 0x07000000
	)

	err := NewError(code, attribute)

	if err.Code != code {
		t.Fatalf("Code = %#08x, want %#08x", err.Code, code)
	}
	if err.Attribute != attribute {
		t.Fatalf("Attribute = %#08x, want %#08x", err.Attribute, attribute)
	}
	if got, want := err.GetCode(), "HMS_0700_0000_0300_800A"; got != want {
		t.Fatalf("GetCode() = %q, want %q", got, want)
	}
}

func TestNewErrorPreservesZeroValues(t *testing.T) {
	err := NewError(0, 0)

	if err.Code != 0 || err.Attribute != 0 {
		t.Fatalf("NewError(0, 0) = %#v, want zero-valued error", err)
	}
	if got := err.GetCode(); got != "" {
		t.Fatalf("GetCode() = %q, want empty string", got)
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name                  string
		code, attr            uint32
		wantCode, wantMessage string
	}{
		{"known", 0x00010006, 0x03000100, "HMS_0300_0100_0001_0006", "The heatbed temperature is abnormal; the sensor may have a short circuit."},
		{"unknown", 0xFFFFFFFF, 0xFFFFFFFF, "HMS_FFFF_FFFF_FFFF_FFFF", "HMS_FFFF_FFFF_FFFF_FFFF"},
		{"zero code", 0, 1, "", ""},
		{"zero attribute", 1, 0, "", ""},
		{"leading zeroes", 1, 1, "HMS_0000_0001_0000_0001", "HMS_0000_0001_0000_0001"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := NewError(tc.code, tc.attr)
			if got := err.GetCode(); got != tc.wantCode {
				t.Errorf("GetCode() = %q, want %q", got, tc.wantCode)
			}
			if got := err.Error(); got != tc.wantMessage {
				t.Errorf("Error() = %q, want %q", got, tc.wantMessage)
			}
		})
	}
}
