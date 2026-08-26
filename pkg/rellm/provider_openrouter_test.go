package rellm

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/or_gemini_tools_req_long_conversation.json
var goldenOrLongConv []byte

//go:embed testdata/image_resp.json
var goldenImageResp []byte

func TestOpenRouterToConversationElements_ImageGeneration(t *testing.T) {
	p := &OpenRouterProvider{}
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

func TestOpenRouterToProviderRepresentation_ImageGeneration(t *testing.T) {
	p := &OpenRouterProvider{}
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

// TestOpenRouterRoundTrip_ImageGeneration_FromFile verifies faithful round-trip
// of the image_generation_call item from the embedded image response fixture.
func TestOpenRouterRoundTrip_ImageGeneration_FromFile(t *testing.T) {
	var resp struct {
		Output []json.RawMessage `json:"output"`
	}
	err := json.Unmarshal(goldenImageResp, &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Output)

	p := &OpenRouterProvider{}
	elems, err := p.ToConversationElements(resp.Output)
	assert.NoError(t, err)
	assert.Len(t, elems, len(resp.Output))

	wireBack, err := p.ToProviderRepresentation(elems)
	assert.NoError(t, err)
	assert.Len(t, wireBack, len(resp.Output))

	for i, original := range resp.Output {
		assert.JSONEq(t, string(original), string(wireBack[i]), "image item %d round-trip mismatch", i)
	}
}

func TestOpenRouterToConversationElements_AssistantMessage(t *testing.T) {
	p := &OpenRouterProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"role":"assistant","content":[{"type":"input_text","text":"hello world"}]}`),
	}
	elems, err := p.ToConversationElements(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	msg, ok := elems[0].(*AssistantMessage)
	assert.True(t, ok)
	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "hello world", TextFromContent(msg.Content))
}

func TestOpenRouterToConversationElements_FunctionCall(t *testing.T) {
	p := &OpenRouterProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"type":"function_call","name":"search","call_id":"abc","arguments":{"query":"test"}}`),
	}
	elems, err := p.ToConversationElements(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	fn, ok := elems[0].(*FunctionCall)
	assert.True(t, ok)
	assert.Equal(t, "search", fn.Name)
	assert.Equal(t, `{"query":"test"}`, string(fn.Args))
}

func TestOpenRouterReasoningContentRoundTrip(t *testing.T) {
	p := &OpenRouterProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"id":"rs-1","type":"reasoning","status":"completed","content":[{"type":"reasoning_text","text":"first"},{"type":"reasoning_text","text":"second"}],"summary":[],"signature":"replay-me"}`),
	}

	elements, err := p.ToConversationElements(items)
	assert.NoError(t, err)
	assert.Len(t, elements, 1)

	reasoning, ok := elements[0].(*Reasoning)
	assert.True(t, ok)
	assert.Equal(t, "rs-1", reasoning.ID)
	assert.Equal(t, "completed", reasoning.Status)
	assert.Equal(t, "first second", reasoning.Text)
	assert.Equal(t, "replay-me", reasoning.Signature)

	wire, err := p.ToProviderRepresentation(elements)
	assert.NoError(t, err)
	assert.Len(t, wire, 1)
	// Signed reasoning blocks retain their empty summary and signature.
	assert.JSONEq(t, `{"id":"rs-1","type":"reasoning","status":"completed","content":[{"type":"reasoning_text","text":"first second"}],"summary":[],"signature":"replay-me"}`, string(wire[0]))
}

func TestOpenRouterReasoningSignatureOnlyRoundTrip(t *testing.T) {
	p := &OpenRouterProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"id":"rs-1","type":"reasoning","status":"completed","summary":[],"signature":"replay-me"}`),
	}

	elements, err := p.ToConversationElements(items)
	assert.NoError(t, err)
	assert.Len(t, elements, 1)

	wire, err := p.ToProviderRepresentation(elements)
	assert.NoError(t, err)
	assert.Len(t, wire, 1)
	assert.JSONEq(t, string(items[0]), string(wire[0]))
}

func TestOpenRouterToProviderRepresentation_AssistantMessage(t *testing.T) {
	p := &OpenRouterProvider{}
	msg := AssistantMessage{MessageContent{Role: "assistant", Content: []MessagePart{{Type: "input_text", Text: "hello world"}}}}
	raw, err := p.ToProviderRepresentation([]ConversationElement{&msg})
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

func TestOpenRouterToProviderRepresentation_UserStructuredContent(t *testing.T) {
	p := &OpenRouterProvider{}
	msg := UserMessage{MessageContent{Role: "user", Content: []MessagePart{
		{Type: "input_text", Text: "hello"},
		{Type: "image_url", ImageURL: &ImageURL{URL: "https://example.com/img.png"}},
	}}}
	raw, err := p.ToProviderRepresentation([]ConversationElement{&msg})
	assert.NoError(t, err)

	var wire struct {
		Role    string        `json:"role"`
		Content []MessagePart `json:"content"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "user", wire.Role)
	assert.Len(t, wire.Content, 2)
}

func TestOpenRouterToProviderRepresentation_FunctionCall(t *testing.T) {
	p := &OpenRouterProvider{}
	fc := FunctionCall{
		ID:     "fc-1",
		Name:   "search",
		Args:   json.RawMessage(`{"query":"test"}`),
		CallID: "call_abc",
	}
	raw, err := p.ToProviderRepresentation([]ConversationElement{&fc})
	assert.NoError(t, err)

	var wire struct {
		Name   string          `json:"name"`
		Args   json.RawMessage `json:"arguments"`
		CallID string          `json:"call_id"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "search", wire.Name)
	assert.Equal(t, `{"query":"test"}`, string(wire.Args))
	assert.Equal(t, "call_abc", wire.CallID)
}

func TestOpenRouterToProviderRepresentation_EmptyElements(t *testing.T) {
	p := &OpenRouterProvider{}
	raw, err := p.ToProviderRepresentation(nil)
	assert.NoError(t, err)
	assert.Nil(t, raw)
}

func TestNewOpenRouterProvider_ValidArgs(t *testing.T) {
	or, err := NewOpenRouterProvider("test-key", Model("test/model"))
	assert.NoError(t, err)
	assert.Equal(t, Model("test/model"), or.Model())
	assert.Equal(t, "https://openrouter.ai/api/v1/responses", or.URL())
	assert.Equal(t, "Bearer test-key", or.Header().Get("Authorization"))
}

func TestNewOpenRouterProvider_EmptyModel(t *testing.T) {
	_, err := NewOpenRouterProvider("test-key", "")
	assert.ErrorIs(t, err, ErrEndpointMissingModelName)
}

func TestNewOpenRouterProvider_EmptyAPIKey(t *testing.T) {
	_, err := NewOpenRouterProvider("", Model("test/model"))
	assert.ErrorIs(t, err, ErrMissingApiKeyForProvider)
}

// TestOpenRouterRoundTrip_FromFile verifies faithful ToConversationElements→ToProviderRepresentation round-trip
// using the embedded golden conversation file. With role-typed messages, each
// type serializes deterministically, so the wire output matches the input byte-for-byte.
func TestOpenRouterRoundTrip_FromFile(t *testing.T) {
	req := struct {
		Model string            `json:"model"`
		Input []json.RawMessage `json:"input"`
	}{}
	err := json.Unmarshal(goldenOrLongConv, &req)
	assert.NoError(t, err)

	p := &OpenRouterProvider{}
	elements, err := p.ToConversationElements(req.Input)
	assert.NoError(t, err)
	assert.Len(t, elements, len(req.Input))

	wireBack, err := p.ToProviderRepresentation(elements)
	assert.NoError(t, err)
	assert.Len(t, wireBack, len(req.Input))

	for i, original := range req.Input {
		assert.JSONEq(t, string(original), string(wireBack[i]), "item %d round-trip mismatch", i)
	}
}

func TestNewOpenRouterProvider_MissingClient(t *testing.T) {
	_, err := NewOpenRouterProviderWithHTTPClient("test-key", "test/model", nil)
	assert.ErrorIs(t, err, ErrMissingHTTPClientForProvider)
}
