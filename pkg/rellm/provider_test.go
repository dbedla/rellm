package rellm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementTypeInterface(t *testing.T) {
	elems := []ConversationElement{
		&UserMessage{},
		&AssistantMessage{},
		&SystemMessage{},
		&FunctionCall{},
		&FunctionCallResponse{},
		&Reasoning{},
		&ImageGeneration{},
	}
	want := []ElementType{
		ElementTypeMessage, ElementTypeMessage, ElementTypeMessage,
		ElementTypeFunctionCall, ElementTypeFunctionCallResp,
		ElementTypeReasoning, ElementTypeImageGeneration,
	}
	for i, e := range elems {
		assert.Equal(t, want[i], e.elementType())
	}
}

func TestTextFromContent(t *testing.T) {
	parts := []MessagePart{
		{Type: "input_text", Text: "hello"},
		{Type: "input_text", Text: ""},
		{Type: "input_text", Text: "world"},
	}
	assert.Equal(t, "hello world", TextFromContent(parts))

	assert.Empty(t, TextFromContent(nil))
}
