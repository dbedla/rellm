package rellm

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/lms/unknown_fn_call_03_req.json
var goldenLMStudioRequest []byte

func TestLMStudioToConversationElements_UserMessage(t *testing.T) {
	p := &LMStudioProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"role":"user","content":[{"type":"input_text","text":"hi"}]}`),
	}
	elems, err := p.ToConversationElements(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	msg, ok := elems[0].(*UserMessage)
	assert.True(t, ok)
	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "hi", TextFromContent(msg.Content))
}

func TestLMStudioToConversationElements_Unknown(t *testing.T) {
	p := &LMStudioProvider{}
	tests := []struct {
		name     string
		raw      json.RawMessage
		typeName string
		role     string
	}{
		{name: "type", raw: json.RawMessage(`{"type":"future_output","value":42}`), typeName: "future_output"},
		{name: "role", raw: json.RawMessage(`{"type":"message","role":"critic","content":"review"}`), typeName: "message", role: "critic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := p.ToConversationElements([]json.RawMessage{tt.raw})
			assert.NoError(t, err)
			assert.Len(t, elements, 1)

			unknown, ok := elements[0].(*UnknownElement)
			assert.True(t, ok)
			assert.Equal(t, ProviderLMStudio, unknown.Provider)
			assert.Equal(t, tt.typeName, unknown.Type)
			assert.Equal(t, tt.role, unknown.Role)
			assert.JSONEq(t, string(tt.raw), string(unknown.Raw))

			wire, err := p.ToProviderRepresentation(elements)
			assert.NoError(t, err)
			assert.Len(t, wire, 1)
			assert.JSONEq(t, string(tt.raw), string(wire[0]))
		})
	}
}

func TestLMStudioToProviderRepresentation_UnknownFromOtherProvider(t *testing.T) {
	p := &LMStudioProvider{}
	element := newUnknownElement(ProviderOpenRouter, "future_output", "", json.RawMessage(`{"type":"future_output"}`))

	wire, err := p.ToProviderRepresentation([]ConversationElement{element})
	assert.ErrorIs(t, err, ErrUnknownElementProviderMismatch)
	assert.Nil(t, wire)
}

func TestLMStudioToConversationElements_ImageGeneration(t *testing.T) {
	p := &LMStudioProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"id":"ig_tmp_vqotwoa5eg","type":"image_generation_call","status":"completed","result":"data:image/jpeg;base64,/9j/4AAQ"}`),
	}
	elems, err := p.ToConversationElements(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	ig, ok := elems[0].(*ImageGeneration)
	assert.True(t, ok)
	assert.Equal(t, "ig_tmp_vqotwoa5eg", ig.ID)
	assert.Equal(t, "completed", ig.Status)
	assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQ", ig.Result)
}

func TestLMStudioToProviderRepresentation_ImageGeneration(t *testing.T) {
	p := &LMStudioProvider{}
	ig := ImageGeneration{ID: "ig_tmp_vqotwoa5eg", Status: "completed", Result: "data:image/jpeg;base64,/9j/4AAQ"}
	raw, err := p.ToProviderRepresentation([]ConversationElement{&ig})
	assert.NoError(t, err)
	assert.Len(t, raw, 1)

	var wire struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "ig_tmp_vqotwoa5eg", wire.ID)
	assert.Equal(t, "image_generation_call", wire.Type)
	assert.Equal(t, "completed", wire.Status)
	assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQ", wire.Result)
}

func TestLMStudioToProviderRepresentation_UserMessage(t *testing.T) {
	p := &LMStudioProvider{}
	msg := UserMessage{MessageContent{Role: "user", Content: []MessagePart{{Type: "input_text", Text: "hi"}}}}
	raw, err := p.ToProviderRepresentation([]ConversationElement{&msg})
	assert.NoError(t, err)
	assert.Len(t, raw, 1)

	var wire map[string]interface{}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "user", wire["role"])
}

func TestLMStudioToProviderRepresentation_EmptyElements(t *testing.T) {
	p := &LMStudioProvider{}
	raw, err := p.ToProviderRepresentation(nil)
	assert.NoError(t, err)
	assert.Nil(t, raw)
}

func TestNewLMStudioProvider_ValidArgs(t *testing.T) {
	lm, err := NewLMStudioProvider(Model("local/model"), "http://localhost", "1234")
	assert.NoError(t, err)
	assert.Equal(t, Model("local/model"), lm.model)
	assert.Equal(t, "http://localhost:1234/v1/responses", lm.url)
}

func TestNewLMStudioProvider_RemoteHost(t *testing.T) {
	lm, err := NewLMStudioProvider(Model("local/model"), "http://192.168.1.50", "1234")
	assert.NoError(t, err)
	assert.Equal(t, "http://192.168.1.50:1234/v1/responses", lm.url)
}

func TestNewLMStudioProvider_MissingModel(t *testing.T) {
	_, err := NewLMStudioProvider("", "localhost", "1234")
	assert.ErrorIs(t, err, ErrEndpointMissingModelName)
}

func TestNewLMStudioProvider_MissingHost(t *testing.T) {
	_, err := NewLMStudioProvider(Model("local/model"), "", "1234")
	assert.ErrorIs(t, err, ErrEndpointMissingHost)
}

func TestNewLMStudioProvider_MissingPort(t *testing.T) {
	_, err := NewLMStudioProvider(Model("local/model"), "localhost", "")
	assert.ErrorIs(t, err, ErrEndpointMissingPort)
}

func TestLMStudioRoundTrip_FromFile(t *testing.T) {
	var req struct {
		Input []json.RawMessage `json:"input"`
	}
	err := json.Unmarshal(goldenLMStudioRequest, &req)
	assert.NoError(t, err)

	p := &LMStudioProvider{}
	elements, err := p.ToConversationElements(req.Input)
	assert.NoError(t, err)
	assert.Len(t, elements, len(req.Input))

	wireBack, err := p.ToProviderRepresentation(elements)
	assert.NoError(t, err)

	// Round-trip should preserve all items.
	assert.Equal(t, len(req.Input), len(wireBack))

	for i, original := range req.Input {
		assert.JSONEq(t, string(original), string(wireBack[i]), "item %d round-trip mismatch", i)
	}
}

func TestNewLMStudioProviderHTTPHost(t *testing.T) {
	p, err := NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:1234/v1/responses", p.url)
}

func TestNewLMStudioProviderHTTPSHost(t *testing.T) {
	p, err := NewLMStudioProvider("google/gemma-4-26b-a4b", "https://127.0.0.1", "1234")
	assert.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1:1234/v1/responses", p.url)
}

func TestNewLMStudioProviderNoHTTPHost(t *testing.T) {
	p, err := NewLMStudioProvider("google/gemma-4-26b-a4b", "127.0.0.1", "1234")
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:1234/v1/responses", p.url)
}

func TestNewLMStudioProvider_MissingClient(t *testing.T) {
	_, err := NewLMStudioProviderWithHTTPClient(Model("local/model"), "localhost", "123", nil)
	assert.ErrorIs(t, err, ErrMissingHTTPClientForProvider)
}
