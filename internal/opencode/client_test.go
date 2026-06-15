package opencode

import "testing"

func TestParseJSONEventsToleratesUnknownEvents(t *testing.T) {
	raw := `{"type":"unknown","payload":{"x":1}}
{"type":"message","text":"Esta es una respuesta final del agente."}
not-json
`
	events, text := ParseJSONEvents(raw)
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}
	if text != "Esta es una respuesta final del agente." {
		t.Fatalf("text = %q", text)
	}
}
