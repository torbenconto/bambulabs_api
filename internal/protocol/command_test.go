package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewCommand(t *testing.T) {
	payload, err := NewCommand(Print).Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"print": map[string]any{
			"sequence_id": "0",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Marshal() = %#v, want %#v", got, want)
	}
}

func TestCommandMarshal(t *testing.T) {
	payload, err := NewCommand(Print).
		WithCommand("gcode_line").
		WithParam("G28").
		Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"print": map[string]any{
			"command":     "gcode_line",
			"param":       "G28",
			"sequence_id": "0",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Marshal() = %#v, want %#v", got, want)
	}
}

func TestSetOverwritesExistingValue(t *testing.T) {
	payload, err := NewCommand(System).
		Set("foo", 1).
		Set("foo", 2).
		Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"system": map[string]any{
			"sequence_id": "0",
			"foo":         float64(2), // JSON numbers decode as float64
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Marshal() = %#v, want %#v", got, want)
	}
}

func TestCustomMessageTypeMarshal(t *testing.T) {
	payload, err := NewCommand(MessageType("custom")).
		Set("foo", "bar").
		Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := map[string]any{
		"custom": map[string]any{
			"sequence_id": "0",
			"foo":         "bar",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Marshal() = %#v, want %#v", got, want)
	}
}
