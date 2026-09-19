package rellm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextFromContent(t *testing.T) {
	parts := []MessagePart{
		{Type: "input_text", Text: "hello"},
		{Type: "input_text", Text: ""},
		{Type: "input_text", Text: "world"},
	}
	assert.Equal(t, "hello world", TextFromContent(parts))

	assert.Empty(t, TextFromContent(nil))
}

type bogusElement struct{}

func (bogusElement) Kind() ElementKind          { return ElementKind("bogus") }
func (bogusElement) Clone() ConversationElement { return bogusElement{} }

func TestStdToProviderRepresentation_UnsupportedElement(t *testing.T) {
	wire, err := StdToProviderRepresentation([]ConversationElement{bogusElement{}})
	assert.Nil(t, wire)
	assert.ErrorIs(t, err, ErrMarshalingConversationElement)
}

func TestConversationElements_MarshalParseRoundTrip(t *testing.T) {
	elements := []ConversationElement{
		&UserMessage{MessageContent: MessageContent{Role: "user", Content: []MessagePart{{Type: "input_text", Text: "hi"}}}},
		&AssistantMessage{MessageContent: MessageContent{Role: "assistant", Content: []MessagePart{{Type: "output_text", Text: "hello"}}}},
		&SystemMessage{MessageContent: MessageContent{Role: "system", Content: []MessagePart{{Type: "input_text", Text: "system prompt"}}}},
		&FunctionCall{ID: "1", Name: "test", CallID: "call_1", Args: json.RawMessage(`{}`)},
		&FunctionCallResp{ID: "2", CallID: "call_1", Output: "ok"},
		&Reasoning{ID: "3", Text: "thinking"},
		&ImageGeneration{ID: "4", Result: "img_id"},
	}
	for _, el := range elements {
		t.Run(string(el.Kind()), func(t *testing.T) {
			b, err := json.Marshal(el)
			assert.NoError(t, err)
			parsed, err := ParseConversationElement(b)
			assert.NoError(t, err)
			assert.Equal(t, el.Kind(), parsed.Kind())
		})
	}
}
