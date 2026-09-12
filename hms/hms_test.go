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
