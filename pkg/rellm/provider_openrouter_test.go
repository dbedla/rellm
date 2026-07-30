package rellm

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/or_gemini_tools_req_long_conversation.json
var goldenOrLongConv []byte

func TestOpenRouterFromWire_Message(t *testing.T) {
	p := &openrouterProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"role":"assistant","content":[{"type":"input_text","text":"hello world"}]}`),
	}
	elems, err := p.FromWire(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	msg, ok := elems[0].(*TextMessage)
	assert.True(t, ok)
	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "hello world", TextFromContent(msg.Content))
}

func TestOpenRouterFromWire_FunctionCall(t *testing.T) {
	p := &openrouterProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"type":"function_call","name":"search","call_id":"abc","arguments":{"query":"test"}}`),
	}
	elems, err := p.FromWire(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	fn, ok := elems[0].(*FunctionCall)
	assert.True(t, ok)
	assert.Equal(t, "search", fn.Name)
	assert.Equal(t, `{"query":"test"}`, string(fn.Args))
}

func TestOpenRouterToWire_TextMessage(t *testing.T) {
	p := &openrouterProvider{}
	msg := TextMessage{
		Role:    "assistant",
		Content: []MessagePart{{Type: "input_text", Text: "hello world"}},
	}
	raw, err := p.ToWire([]ConversationElement{&msg})
	assert.NoError(t, err)
	assert.Len(t, raw, 1)

	var wire struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "assistant", wire.Role)
	assert.Equal(t, "hello world", wire.Content)
}

func TestOpenRouterToWire_UserStructuredContent(t *testing.T) {
	p := &openrouterProvider{}
	msg := TextMessage{
		Role:    "user",
		Content: []MessagePart{
			{Type: "input_text", Text: "hello"},
			{Type: "image_url", ImageURL: &ImageURL{URL: "https://example.com/img.png"}},
		},
	}
	raw, err := p.ToWire([]ConversationElement{&msg})
	assert.NoError(t, err)

	var wire struct {
		Role    string         `json:"role"`
		Content []MessagePart  `json:"content"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "user", wire.Role)
	assert.Len(t, wire.Content, 2)
}

func TestOpenRouterToWire_FunctionCall(t *testing.T) {
	p := &openrouterProvider{}
	fc := FunctionCall{
		Id:     "fc-1",
		Name:   "search",
		Args:   json.RawMessage(`{"query":"test"}`),
		CallId: "call_abc",
	}
	raw, err := p.ToWire([]ConversationElement{&fc})
	assert.NoError(t, err)

	var wire struct {
		Name    string          `json:"name"`
		Args    json.RawMessage `json:"arguments"`
		CallId  string          `json:"call_id"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "search", wire.Name)
	assert.Equal(t, `{"query":"test"}`, string(wire.Args))
	assert.Equal(t, "call_abc", wire.CallId)
}

func TestOpenRouterToWire_EmptyElements(t *testing.T) {
	p := &openrouterProvider{}
	raw, err := p.ToWire(nil)
	assert.NoError(t, err)
	assert.Nil(t, raw)
}

func TestNewOpenRouterEndpoint_ValidArgs(t *testing.T) {
	or, err := NewOpenRouterEndpoint("test-key", Model("test/model"))
	assert.NoError(t, err)
	assert.Equal(t, Provider_OpenRouter, or.provider)
	assert.Equal(t, "https://router.openrouter.ai/v1/responses", or.rae.GetUrl())
	assert.Equal(t, "Bearer test-key", or.rae.GetHttpHeader().Get("Authorization"))
}

func TestNewOpenRouterEndpoint_EmptyModel(t *testing.T) {
	_, err := NewOpenRouterEndpoint("test-key", "")
	assert.ErrorIs(t, err, ErrEndpointMissingModelName)
}

func TestNewOpenRouterEndpoint_EmptyApiKey(t *testing.T) {
	_, err := NewOpenRouterEndpoint("", Model("test/model"))
	assert.EqualError(t, err, "missing API key for OpenRouter endpoint")
}

// TestOpenRouterRoundTrip_FromFile verifies FromWire→ToWire round-trip fidelity
// using the embedded golden conversation file.
func TestOpenRouterRoundTrip_FromFile(t *testing.T) {
	req := struct {
		Model string               `json:"model"`
		Input []json.RawMessage    `json:"input"`
	}{}
	err := json.Unmarshal(goldenOrLongConv, &req)
	assert.NoError(t, err)

	p := &openrouterProvider{}
	elements, err := p.FromWire(req.Input)
	assert.NoError(t, err)
	assert.Len(t, elements, len(req.Input))

	wireBack, err := p.ToWire(elements)
	assert.NoError(t, err)

	// Round-trip should preserve all items.
	assert.Equal(t, len(req.Input), len(wireBack))

	for i, original := range req.Input {
		assert.JSONEq(t, string(original), string(wireBack[i]), "item %d round-trip mismatch", i)
	}
}
