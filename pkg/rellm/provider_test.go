package rellm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementTypeInterface(t *testing.T) {
	elems := []ConversationElement{
		&TextMessage{},
		&FunctionCall{},
		&FunctionCallResponse{},
		&Reasoning{},
		&ImageGeneration{},
	}
	want := []ElementType{
		ElementTypeMessage, ElementTypeFunctionCall, ElementTypeFunctionCallResp,
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

func TestProviderRegistry(t *testing.T) {
	for _, prov := range []Provider{Provider_OpenRouter, Provider_LMStudio} {
		cfg, err := getProviderConfig(prov)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	}

	_, err := getProviderConfig("totally_unknown")
	assert.Error(t, err)
}
