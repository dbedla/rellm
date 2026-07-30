package rellm

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/pro_api_tool_B2_req_unknown_fn_call.json
var goldenLMStudioRequest []byte

func TestLMStudioFromWire_UserMessage(t *testing.T) {
	p := &lmstudioProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"role":"user","content":[{"type":"input_text","text":"hi"}]}`),
	}
	elems, err := p.FromWire(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	msg, ok := elems[0].(*UserMessage)
	assert.True(t, ok)
	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "hi", TextFromContent(msg.Content))
}

func TestLMStudioFromWire_ImageGeneration(t *testing.T) {
	p := &lmstudioProvider{}
	items := []json.RawMessage{
		json.RawMessage(`{"id":"ig_tmp_vqotwoa5eg","type":"image_generation_call","status":"completed","result":"data:image/jpeg;base64,/9j/4AAQ"}`),
	}
	elems, err := p.FromWire(items)
	assert.NoError(t, err)
	assert.Len(t, elems, 1)

	ig, ok := elems[0].(*ImageGeneration)
	assert.True(t, ok)
	assert.Equal(t, "ig_tmp_vqotwoa5eg", ig.Id)
	assert.Equal(t, "completed", ig.Status)
	assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQ", ig.Result)
}

func TestLMStudioToWire_ImageGeneration(t *testing.T) {
	p := &lmstudioProvider{}
	ig := ImageGeneration{Id: "ig_tmp_vqotwoa5eg", Status: "completed", Result: "data:image/jpeg;base64,/9j/4AAQ"}
	raw, err := p.ToWire([]ConversationElement{&ig})
	assert.NoError(t, err)
	assert.Len(t, raw, 1)

	var wire struct {
		Id     string `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "ig_tmp_vqotwoa5eg", wire.Id)
	assert.Equal(t, "image_generation_call", wire.Type)
	assert.Equal(t, "completed", wire.Status)
	assert.Equal(t, "data:image/jpeg;base64,/9j/4AAQ", wire.Result)
}

func TestLMStudioToWire_UserMessage(t *testing.T) {
	p := &lmstudioProvider{}
	msg := UserMessage{messageContent{Role: "user", Content: []MessagePart{{Type: "input_text", Text: "hi"}}}}
	raw, err := p.ToWire([]ConversationElement{&msg})
	assert.NoError(t, err)
	assert.Len(t, raw, 1)

	var wire map[string]interface{}
	err = json.Unmarshal(raw[0], &wire)
	assert.NoError(t, err)
	assert.Equal(t, "user", wire["role"])
}

func TestLMStudioToWire_EmptyElements(t *testing.T) {
	p := &lmstudioProvider{}
	raw, err := p.ToWire(nil)
	assert.NoError(t, err)
	assert.Nil(t, raw)
}

func TestNewLMStudioEndpoint_ValidArgs(t *testing.T) {
	lm, err := NewLMStudioEndpoint(Model("local/model"), "localhost", "1234")
	assert.NoError(t, err)
	assert.Equal(t, Provider_LMStudio, lm.provider)
	assert.Equal(t, "http://localhost:1234/v1/responses", lm.rae.GetUrl())
}

func TestNewLMStudioEndpoint_RemoteHost(t *testing.T) {
	lm, err := NewLMStudioEndpoint(Model("local/model"), "192.168.1.50", "1234")
	assert.NoError(t, err)
	assert.Equal(t, "http://192.168.1.50:1234/v1/responses", lm.rae.GetUrl())
}

func TestNewLMStudioEndpoint_MissingModel(t *testing.T) {
	_, err := NewLMStudioEndpoint("", "localhost", "1234")
	assert.ErrorIs(t, err, ErrEndpointMissingModelName)
}

func TestNewLMStudioEndpoint_MissingHost(t *testing.T) {
	_, err := NewLMStudioEndpoint(Model("local/model"), "", "1234")
	assert.ErrorIs(t, err, ErrEndpointMissingHost)
}

func TestNewLMStudioEndpoint_MissingPort(t *testing.T) {
	_, err := NewLMStudioEndpoint(Model("local/model"), "localhost", "")
	assert.ErrorIs(t, err, ErrEndpointMissingPort)
}

func TestLMStudioRoundTrip_FromFile(t *testing.T) {
	var req struct {
		Input []json.RawMessage `json:"input"`
	}
	err := json.Unmarshal(goldenLMStudioRequest, &req)
	assert.NoError(t, err)

	p := &lmstudioProvider{}
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
